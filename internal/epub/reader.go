package epub

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Metadata struct {
	Title       string
	Authors     []string
	Description string
	Publisher   string
	Language    string
	Identifiers map[string]string
	CoverItemID string
	OpfDir      string
}

type containerRoot struct {
	XMLName   xml.Name   `xml:"container"`
	Rootfiles []rootfile `xml:"rootfiles>rootfile"`
}

type rootfile struct {
	FullPath  string `xml:"full-path,attr"`
	MediaType string `xml:"media-type,attr"`
}

type opfPackage struct {
	XMLName  xml.Name    `xml:"package"`
	Metadata opfMetadata `xml:"metadata"`
	Manifest opfManifest `xml:"manifest"`
}

type opfMetadata struct {
	Titles      []string        `xml:"title"`
	Creators    []string        `xml:"creator"`
	Description string          `xml:"description"`
	Publisher   string          `xml:"publisher"`
	Language    string          `xml:"language"`
	MetaTags    []opfMetaTag    `xml:"meta"`
	Identifiers []opfIdentifier `xml:"identifier"`
}

type opfMetaTag struct {
	Name     string `xml:"name,attr"`
	Content  string `xml:"content,attr"`
	Property string `xml:"property,attr"`
	Value    string `xml:",chardata"`
}

type opfIdentifier struct {
	ID     string `xml:"id,attr"`
	Scheme string `xml:"scheme,attr"`
	Value  string `xml:",chardata"`
}

type opfManifest struct {
	Items []opfItem `xml:"item"`
}

type opfItem struct {
	ID         string `xml:"id,attr"`
	Href       string `xml:"href,attr"`
	MediaType  string `xml:"media-type,attr"`
	Properties string `xml:"properties,attr"`
}

// ExtractMetadata extrai metadados do EPUB em milissegundos sem alocar o livro em memória RAM.
func ExtractMetadata(epubPath string) (*Metadata, error) {
	zr, err := zip.OpenReader(epubPath)
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir zip: %w", err)
	}
	defer zr.Close()

	fileMap := make(map[string]*zip.File, len(zr.File))
	for _, f := range zr.File {
		fileMap[f.Name] = f
	}

	// 1. Localizar o ficheiro OPF a partir de META-INF/container.xml
	containerFile, ok := fileMap["META-INF/container.xml"]
	if !ok {
		return nil, errors.New("EPUB inválido: container.xml não encontrado")
	}

	rc, err := containerFile.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	var cont containerRoot
	if err := xml.NewDecoder(rc).Decode(&cont); err != nil {
		return nil, fmt.Errorf("erro ao decodificar container.xml: %w", err)
	}

	if len(cont.Rootfiles) == 0 || cont.Rootfiles[0].FullPath == "" {
		return nil, errors.New("caminho OPF não especificado no container")
	}

	opfPath := cont.Rootfiles[0].FullPath
	opfFile, ok := fileMap[opfPath]
	if !ok {
		return nil, fmt.Errorf("ficheiro OPF não encontrado: %s", opfPath)
	}

	// 2. Decode por streaming do content.opf
	opfRc, err := opfFile.Open()
	if err != nil {
		return nil, err
	}
	defer opfRc.Close()

	var pkg opfPackage
	if err := xml.NewDecoder(opfRc).Decode(&pkg); err != nil {
		return nil, fmt.Errorf("erro no decode do OPF: %w", err)
	}

	meta := &Metadata{
		Description: strings.TrimSpace(pkg.Metadata.Description),
		Publisher:   strings.TrimSpace(pkg.Metadata.Publisher),
		Language:    strings.TrimSpace(pkg.Metadata.Language),
		Identifiers: make(map[string]string),
		OpfDir:      path.Dir(opfPath),
	}

	if len(pkg.Metadata.Titles) > 0 {
		meta.Title = strings.TrimSpace(pkg.Metadata.Titles[0])
	}
	for _, cr := range pkg.Metadata.Creators {
		trimmed := strings.TrimSpace(cr)
		if trimmed != "" {
			meta.Authors = append(meta.Authors, trimmed)
		}
	}
	for _, id := range pkg.Metadata.Identifiers {
		if id.Scheme != "" && id.Value != "" {
			meta.Identifiers[id.Scheme] = strings.TrimSpace(id.Value)
		}
	}

	// 3. Resolução da Capa (EPUB 2 e EPUB 3)
	var coverID string
	for _, m := range pkg.Metadata.MetaTags {
		if m.Name == "cover" {
			coverID = m.Content
			break
		}
	}

	var coverHref string
	for _, it := range pkg.Manifest.Items {
		if strings.Contains(it.Properties, "cover-image") {
			coverHref = it.Href
			break
		}
		if coverID != "" && it.ID == coverID {
			coverHref = it.Href
			break
		}
	}

	if coverHref == "" {
		for _, it := range pkg.Manifest.Items {
			if strings.HasPrefix(it.MediaType, "image/") && strings.Contains(strings.ToLower(it.Href), "cover") {
				coverHref = it.Href
				break
			}
		}
	}

	if coverHref != "" {
		resolvedCoverPath := coverHref
		if meta.OpfDir != "." && meta.OpfDir != "" {
			resolvedCoverPath = path.Join(meta.OpfDir, coverHref)
		}
		meta.CoverItemID = resolvedCoverPath
	}

	return meta, nil
}

// ExtractCover extrai a capa do EPUB escrevendo no io.Writer fornecido.
func ExtractCover(epubPath string, w io.Writer) error {
	meta, err := ExtractMetadata(epubPath)
	if err != nil {
		return err
	}
	if meta.CoverItemID == "" {
		return errors.New("nenhuma imagem de capa detectada no manifesto")
	}

	zr, err := zip.OpenReader(epubPath)
	if err != nil {
		return err
	}
	defer zr.Close()

	var targetZipFile *zip.File
	for _, f := range zr.File {
		if f.Name == meta.CoverItemID || strings.HasSuffix(f.Name, meta.CoverItemID) {
			targetZipFile = f
			break
		}
	}

	if targetZipFile == nil {
		return fmt.Errorf("imagem de capa referenciada (%s) não existe no arquivo", meta.CoverItemID)
	}

	srcStream, err := targetZipFile.Open()
	if err != nil {
		return err
	}
	defer srcStream.Close()

	_, err = io.Copy(w, srcStream)
	return err
}

// ExtractCoverToFile extrai a capa diretamente do ZIP para o disco com zero buffering na memória heap.
func ExtractCoverToFile(epubPath string, destFilePath string) error {
	if err := os.MkdirAll(filepath.Dir(destFilePath), 0755); err != nil {
		return err
	}

	destFile, err := os.Create(destFilePath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	return ExtractCover(epubPath, destFile)
}

