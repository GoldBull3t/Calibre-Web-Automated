package handlers

const cwaSettingsHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Calibre-Web Automated | Settings</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link href="/static/css/libs/bootstrap.min.css" rel="stylesheet" media="screen">
    <link href="/static/css/style.css" rel="stylesheet" media="screen">
    <link href="/static/css/caliBlur.css" rel="stylesheet" media="screen">
    <link href="/static/css/caliBlur_override.css" rel="stylesheet" media="screen">
    <link href="/static/css/layout-refresh.css" rel="stylesheet" media="screen">
    <link href="/static/css/cwa.css?v=2026.5" rel="stylesheet" media="screen">
</head>
<body class="discover blur">
    <!-- Static navbar -->
    <div class="navbar navbar-default navbar-static-top" role="navigation">
        <div class="container-fluid">
            <div class="navbar-header">
                <button type="button" class="navbar-toggle" data-toggle="collapse" data-target=".navbar-collapse">
                    <span class="sr-only">Toggle Navigation</span>
                    <span class="icon-bar"></span>
                    <span class="icon-bar"></span>
                    <span class="icon-bar"></span>
                </button>
                <a class="navbar-brand" href="/"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#38bdf8" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 8px; flex-shrink: 0;"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path></svg>Calibre-Web Automated <span class="label label-warning" style="font-size: 10px; vertical-align: middle; margin-left: 5px;">CWA</span></a>
            </div>
            <div class="home-btn"><a class="home-btn-tooltip" href="/" data-toggle="tooltip" title="" data-placement="bottom" data-original-title="Home"></a></div>
            <div class="plexBack"><a href="/"></a></div>

            <form class="navbar-form navbar-left" role="search" action="/search" method="GET">
                <div class="form-group input-group input-group-sm">
                    <label for="query" class="sr-only">Search</label>
                    <input type="text" class="form-control" id="query" name="query" placeholder="Search Library">
                    <span class="input-group-btn">
                        <button type="submit" id="query_submit" class="btn btn-default"><span class="glyphicon glyphicon-search"></span></button>
                    </span>
                </div>
            </form>

            <div class="navbar-collapse collapse">
                <ul class="nav navbar-nav">
                    <li><a href="/advsearch" id="advanced_search"><span class="glyphicon glyphicon-search" style="padding-left: 10px; padding-right: 6px;"></span><span class="hidden-sm"> Advanced Search</span></a></li>
                </ul>
                <ul class="nav navbar-nav navbar-right" id="main-nav">
                    <li>
                        <form id="form-upload" class="navbar-form" action="/upload" method="post" enctype="multipart/form-data">
                            <div class="form-group">
                                <span class="btn btn-default btn-file"><span class="glyphicon glyphicon-upload"></span> Upload
                                    <input id="btn-upload" name="file" type="file" accept=".epub,.pdf,.mobi,.azw3,.cbr,.cbz" onchange="this.form.submit()">
                                </span>
                            </div>
                        </form>
                    </li>
                    <li class="top_tasks"><a id="top_tasks" href="/tasks"><span class="glyphicon glyphicon-tasks"></span> <span class="hidden-sm">Tasks</span></a></li>
                    <li><a id="top_admin" href="/admin"><span class="glyphicon glyphicon-dashboard"></span> <span class="hidden-sm">Admin</span></a></li>
                    <li class="cwa-switch-theme">
                        <a id="cwa-switch-theme" href="/cwa-switch-theme">
                            <span class="glyphicon glyphicon-sunglasses"></span>
                            <span class="hidden-sm">Switch Theme</span>
                        </a>
                    </li>
                    <li><a id="refresh-library" style="cursor: pointer;" href="javascript:void(0)" onclick="fetch('/cwa-library-refresh',{method:'POST'}).then(function(){location.reload();});"><span class="glyphicon glyphicon-refresh"></span> <span class="hidden-sm">Refresh Library</span></a></li>
                    <li class="dropdown">
                        <a href="#" class="dropdown-toggle profileDrop" data-toggle="dropdown" role="button" aria-haspopup="true" aria-expanded="false"><span class="glyphicon glyphicon-user"></span></a>
                        <ul class="dropdown-menu profileDropli">
                            <li><a id="top_user" href="/me"><span class="glyphicon glyphicon-user"></span> <span class="hidden-sm">Profile</span></a></li>
                            <li><a href="/cwa-settings"><span class="glyphicon glyphicon-cog"></span> CWA Settings</a></li>
                            <li><a href="/cwa-stats"><span class="glyphicon glyphicon-stats"></span> Deep Analytics</a></li>
                            <li><a href="/duplicates"><span class="glyphicon glyphicon-duplicate"></span> Duplicates</a></li>
                            <li><a href="/cwa-backup/list"><span class="glyphicon glyphicon-floppy-save"></span> Backups</a></li>
                            <li role="separator" class="divider"></li>
                            <li><a href="/logout"><span class="glyphicon glyphicon-log-out"></span> Logout</a></li>
                        </ul>
                    </li>
                </ul>
            </div>
        </div>
    </div>
    <div class="cwa-container">
        <div class="cwa-page-header">
            <div>
                <a href="/" class="btn btn-default cwa-back-btn" style="margin-bottom: 12px;">
                    <span class="glyphicon glyphicon-arrow-left"></span> Back to Library
                </a>
                <h1 class="cwa-page-title">
                    <span class="glyphicon glyphicon-cog" style="color: #38bdf8;"></span>
                    Calibre-Web Automated Settings
                </h1>
                <p class="cwa-page-subtitle">Configure automated background services, conversion targets, deduplication, and snapshot retention.</p>
            </div>
            <div>
                <button type="button" class="btn btn-primary" onclick="saveSettings()">
                    <span class="glyphicon glyphicon-ok"></span> Save Changes
                </button>
            </div>
        </div>

        <div id="save-alert" class="cwa-alert cwa-alert-success" style="display: none;">
            <span class="glyphicon glyphicon-ok-sign"></span> Settings have been saved successfully.
        </div>

        <!-- Automated Services Configuration -->
        <div class="cwa-panel">
            <h2 class="cwa-panel-title">
                <span class="glyphicon glyphicon-tasks"></span>
                Automated Background Services
            </h2>
            
            <div class="cwa-setting-row">
                <div class="cwa-setting-info">
                    <div class="cwa-setting-title">
                        <span class="glyphicon glyphicon-import"></span>
                        Automatic Ingest Service
                    </div>
                    <div class="cwa-setting-desc">Continuously monitors the ingestion directory and automatically imports books across 27 supported formats.</div>
                </div>
                <div class="cwa-setting-control">
                    <label class="cwa-switch" for="enable_ingest">
                        <input type="checkbox" id="enable_ingest" checked>
                        <span class="cwa-slider"></span>
                    </label>
                </div>
            </div>

            <div class="cwa-setting-row">
                <div class="cwa-setting-info">
                    <div class="cwa-setting-title">
                        <span class="glyphicon glyphicon-refresh"></span>
                        Automatic Conversion Service
                    </div>
                    <div class="cwa-setting-desc">Automatically executes format conversion on newly ingested files to match your configured target format.</div>
                </div>
                <div class="cwa-setting-control">
                    <label class="cwa-switch" for="enable_conversion">
                        <input type="checkbox" id="enable_conversion" checked>
                        <span class="cwa-slider"></span>
                    </label>
                </div>
            </div>

            <div class="cwa-setting-row">
                <div class="cwa-setting-info">
                    <div class="cwa-setting-title">
                        <span class="glyphicon glyphicon-picture"></span>
                        Automatic Enforcement of Covers &amp; Metadata
                    </div>
                    <div class="cwa-setting-desc">Directly updates internal EPUB ZIP archives and OPF files so connected e-readers display updated covers and synopsis.</div>
                </div>
                <div class="cwa-setting-control">
                    <label class="cwa-switch" for="enable_enforce">
                        <input type="checkbox" id="enable_enforce" checked>
                        <span class="cwa-slider"></span>
                    </label>
                </div>
            </div>

            <div class="cwa-setting-row">
                <div class="cwa-setting-info">
                    <div class="cwa-setting-title">
                        <span class="glyphicon glyphicon-wrench"></span>
                        Automatic EPUB Fixer Service
                    </div>
                    <div class="cwa-setting-desc">Repairs corrupt archives, validates BCP-47 language tags, and sanitizes XML syntax to prevent Send-to-Kindle rejections.</div>
                </div>
                <div class="cwa-setting-control">
                    <label class="cwa-switch" for="enable_epub_fixer">
                        <input type="checkbox" id="enable_epub_fixer" checked>
                        <span class="cwa-slider"></span>
                    </label>
                </div>
            </div>

            <div class="cwa-setting-row">
                <div class="cwa-setting-info">
                    <div class="cwa-setting-title">
                        <span class="glyphicon glyphicon-compressed"></span>
                        Auto-Compression of Backed Up Files (Auto-Zipper)
                    </div>
                    <div class="cwa-setting-desc">Compresses processed original files in Deflate format to minimize library storage usage.</div>
                </div>
                <div class="cwa-setting-control">
                    <label class="cwa-switch" for="enable_auto_zipper">
                        <input type="checkbox" id="enable_auto_zipper" checked>
                        <span class="cwa-slider"></span>
                    </label>
                </div>
            </div>

            <div class="cwa-setting-row">
                <div class="cwa-setting-info">
                    <div class="cwa-setting-title">
                        <span class="glyphicon glyphicon-tag"></span>
                        Automatic Hardcover ID Fetch
                    </div>
                    <div class="cwa-setting-desc">Queries the Hardcover GraphQL API to automatically resolve and populate missing Hardcover book identifiers.</div>
                </div>
                <div class="cwa-setting-control">
                    <label class="cwa-switch" for="enable_hardcover">
                        <input type="checkbox" id="enable_hardcover" checked>
                        <span class="cwa-slider"></span>
                    </label>
                </div>
            </div>
        </div>

        <!-- Conversion Preferences -->
        <div class="cwa-panel">
            <h2 class="cwa-panel-title">
                <span class="glyphicon glyphicon-transfer"></span>
                Conversion Engine Preferences
            </h2>
            <div class="row">
                <div class="col-sm-6">
                    <div class="cwa-form-group">
                        <label for="target_format" class="cwa-label">Default Target Format:</label>
                        <select id="target_format" class="form-control cwa-select">
                            <option value="EPUB" selected>EPUB (Standard Open Format)</option>
                            <option value="KEPUB">KEPUB (Optimized for Kobo e-Readers)</option>
                            <option value="MOBI">MOBI (Legacy Kindle Format)</option>
                            <option value="AZW3">AZW3 (Modern KF8 Kindle Format)</option>
                            <option value="PDF">PDF (Portable Document Format)</option>
                        </select>
                        <p class="cwa-page-subtitle" style="margin-top: 6px;">Uses kepubify or ebook-convert based on selected format.</p>
                    </div>
                </div>
                <div class="col-sm-6">
                    <div class="cwa-form-group" style="padding-top: 25px;">
                        <label class="cwa-switch" for="retain_processed" style="margin-right: 10px;">
                            <input type="checkbox" id="retain_processed" checked>
                            <span class="cwa-slider"></span>
                        </label>
                        <span style="color: #cbd5e1; font-size: 13px;">Retain original files in <code>processed_books</code></span>
                    </div>
                </div>
            </div>
        </div>

        <!-- Duplicate Detection & Deduplication -->
        <div class="cwa-panel">
            <h2 class="cwa-panel-title">
                <span class="glyphicon glyphicon-duplicate"></span>
                Duplicate Detection &amp; Resolution
            </h2>
            <div class="row">
                <div class="col-sm-6">
                    <div class="cwa-form-group">
                        <label for="dup_strategy" class="cwa-label">Default Auto-Resolution Strategy:</label>
                        <select id="dup_strategy" class="form-control cwa-select">
                            <option value="newest" selected>Keep Newest (Retain recent copy, merge formats)</option>
                            <option value="oldest">Keep Oldest (Retain earliest added copy)</option>
                            <option value="merge">Merge Formats (Combine all unique format files)</option>
                            <option value="most_metadata">Keep Most Metadata (Retain entry with richest tags and synopsis)</option>
                        </select>
                    </div>
                </div>
                <div class="col-sm-6">
                    <div class="cwa-form-group">
                        <label for="dup_threshold" class="cwa-label">Fuzzy Match Sensitivity:</label>
                        <select id="dup_threshold" class="form-control cwa-select">
                            <option value="90" selected>90% Similarity (Strict)</option>
                            <option value="85">85% Similarity (Balanced)</option>
                            <option value="80">80% Similarity (Broad)</option>
                        </select>
                    </div>
                </div>
            </div>
        </div>

        <!-- Automated Backup Service -->
        <div class="cwa-panel">
            <h2 class="cwa-panel-title">
                <span class="glyphicon glyphicon-floppy-save"></span>
                Automated Backup Service
            </h2>
            <div class="row">
                <div class="col-sm-6">
                    <div class="cwa-form-group">
                        <label for="backup_retention" class="cwa-label">Snapshot Retention Limit:</label>
                        <select id="backup_retention" class="form-control cwa-select">
                            <option value="5">Keep 5 Snapshots</option>
                            <option value="10" selected>Keep 10 Snapshots</option>
                            <option value="20">Keep 20 Snapshots</option>
                            <option value="50">Keep 50 Snapshots</option>
                        </select>
                        <p class="cwa-page-subtitle" style="margin-top: 6px;">Older compressed .zip snapshots are auto-pruned to conserve disk space.</p>
                    </div>
                </div>
                <div class="col-sm-6" style="padding-top: 25px;">
                    <button type="button" class="btn btn-default" onclick="fetch('/cwa-backup/create',{method:'POST'}).then(r=>r.json()).then(d=>alert(d.message||'Backup created.'))">
                        <span class="glyphicon glyphicon-save"></span> Trigger Instant Snapshot Now
                    </button>
                </div>
            </div>
        </div>
    </div>

    <script src="/static/js/libs/jquery.min.js"></script>
    <script src="/static/js/libs/bootstrap.min.js"></script>
    <script>
    function saveSettings() {
        fetch('/cwa-settings', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({
                enable_ingest: document.getElementById('enable_ingest').checked,
                enable_conversion: document.getElementById('enable_conversion').checked,
                enable_enforce: document.getElementById('enable_enforce').checked,
                enable_epub_fixer: document.getElementById('enable_epub_fixer').checked,
                target_format: document.getElementById('target_format').value
            })
        }).then(r => {
            const el = document.getElementById('save-alert');
            el.style.display = 'flex';
            setTimeout(() => { el.style.display = 'none'; }, 3000);
        });
    }
    </script>
</body>
</html>`

const cwaConvertHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Calibre-Web Automated | Convert Library</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link href="/static/css/libs/bootstrap.min.css" rel="stylesheet" media="screen">
    <link href="/static/css/style.css" rel="stylesheet" media="screen">
    <link href="/static/css/caliBlur.css" rel="stylesheet" media="screen">
    <link href="/static/css/caliBlur_override.css" rel="stylesheet" media="screen">
    <link href="/static/css/layout-refresh.css" rel="stylesheet" media="screen">
    <link href="/static/css/cwa.css?v=2026.5" rel="stylesheet" media="screen">
</head>
<body class="discover blur">
    <!-- Static navbar -->
    <div class="navbar navbar-default navbar-static-top" role="navigation">
        <div class="container-fluid">
            <div class="navbar-header">
                <button type="button" class="navbar-toggle" data-toggle="collapse" data-target=".navbar-collapse">
                    <span class="sr-only">Toggle Navigation</span>
                    <span class="icon-bar"></span>
                    <span class="icon-bar"></span>
                    <span class="icon-bar"></span>
                </button>
                <a class="navbar-brand" href="/"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#38bdf8" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 8px; flex-shrink: 0;"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path></svg>Calibre-Web Automated <span class="label label-warning" style="font-size: 10px; vertical-align: middle; margin-left: 5px;">CWA</span></a>
            </div>
            <div class="home-btn"><a class="home-btn-tooltip" href="/" data-toggle="tooltip" title="" data-placement="bottom" data-original-title="Home"></a></div>
            <div class="plexBack"><a href="/"></a></div>

            <form class="navbar-form navbar-left" role="search" action="/search" method="GET">
                <div class="form-group input-group input-group-sm">
                    <label for="query" class="sr-only">Search</label>
                    <input type="text" class="form-control" id="query" name="query" placeholder="Search Library">
                    <span class="input-group-btn">
                        <button type="submit" id="query_submit" class="btn btn-default"><span class="glyphicon glyphicon-search"></span></button>
                    </span>
                </div>
            </form>

            <div class="navbar-collapse collapse">
                <ul class="nav navbar-nav">
                    <li><a href="/advsearch" id="advanced_search"><span class="glyphicon glyphicon-search" style="padding-left: 10px; padding-right: 6px;"></span><span class="hidden-sm"> Advanced Search</span></a></li>
                </ul>
                <ul class="nav navbar-nav navbar-right" id="main-nav">
                    <li>
                        <form id="form-upload" class="navbar-form" action="/upload" method="post" enctype="multipart/form-data">
                            <div class="form-group">
                                <span class="btn btn-default btn-file"><span class="glyphicon glyphicon-upload"></span> Upload
                                    <input id="btn-upload" name="file" type="file" accept=".epub,.pdf,.mobi,.azw3,.cbr,.cbz" onchange="this.form.submit()">
                                </span>
                            </div>
                        </form>
                    </li>
                    <li class="top_tasks"><a id="top_tasks" href="/tasks"><span class="glyphicon glyphicon-tasks"></span> <span class="hidden-sm">Tasks</span></a></li>
                    <li><a id="top_admin" href="/admin"><span class="glyphicon glyphicon-dashboard"></span> <span class="hidden-sm">Admin</span></a></li>
                    <li class="cwa-switch-theme">
                        <a id="cwa-switch-theme" href="/cwa-switch-theme">
                            <span class="glyphicon glyphicon-sunglasses"></span>
                            <span class="hidden-sm">Switch Theme</span>
                        </a>
                    </li>
                    <li><a id="refresh-library" style="cursor: pointer;" href="javascript:void(0)" onclick="fetch('/cwa-library-refresh',{method:'POST'}).then(function(){location.reload();});"><span class="glyphicon glyphicon-refresh"></span> <span class="hidden-sm">Refresh Library</span></a></li>
                    <li class="dropdown">
                        <a href="#" class="dropdown-toggle profileDrop" data-toggle="dropdown" role="button" aria-haspopup="true" aria-expanded="false"><span class="glyphicon glyphicon-user"></span></a>
                        <ul class="dropdown-menu profileDropli">
                            <li><a id="top_user" href="/me"><span class="glyphicon glyphicon-user"></span> <span class="hidden-sm">Profile</span></a></li>
                            <li><a href="/cwa-settings"><span class="glyphicon glyphicon-cog"></span> CWA Settings</a></li>
                            <li><a href="/cwa-stats"><span class="glyphicon glyphicon-stats"></span> Deep Analytics</a></li>
                            <li><a href="/duplicates"><span class="glyphicon glyphicon-duplicate"></span> Duplicates</a></li>
                            <li><a href="/cwa-backup/list"><span class="glyphicon glyphicon-floppy-save"></span> Backups</a></li>
                            <li role="separator" class="divider"></li>
                            <li><a href="/logout"><span class="glyphicon glyphicon-log-out"></span> Logout</a></li>
                        </ul>
                    </li>
                </ul>
            </div>
        </div>
    </div>
    <div class="cwa-container">
        <div class="cwa-page-header">
            <div>
                <a href="/" class="btn btn-default cwa-back-btn" style="margin-bottom: 12px;">
                    <span class="glyphicon glyphicon-arrow-left"></span> Back to Library
                </a>
                <h1 class="cwa-page-title">
                    <span class="glyphicon glyphicon-refresh" style="color: #38bdf8;"></span>
                    Convert Library Tool
                </h1>
                <p class="cwa-page-subtitle">Batch multi-format conversion engine for whole-library format standardization.</p>
            </div>
            <div>
                <span class="cwa-badge cwa-badge-info">Multi-Format Engine</span>
            </div>
        </div>

        <div class="cwa-panel">
            <h2 class="cwa-panel-title">
                <span class="glyphicon glyphicon-cog"></span>
                Conversion Execution
            </h2>
            <p class="cwa-panel-description">
                Scans your library collection and performs background format conversion on books that lack the target format.
                Original files are backed up safely in <code>processed_books</code>.
            </p>

            <div class="row">
                <div class="col-sm-4">
                    <div class="cwa-form-group">
                        <label for="convert_target" class="cwa-label">Target Format:</label>
                        <select id="convert_target" class="form-control cwa-select">
                            <option value="EPUB" selected>EPUB (Universal)</option>
                            <option value="KEPUB">KEPUB (Kobo Enhanced)</option>
                            <option value="MOBI">MOBI (Legacy Kindle)</option>
                            <option value="AZW3">AZW3 (KF8)</option>
                            <option value="PDF">PDF</option>
                        </select>
                    </div>
                </div>
            </div>

            <div class="cwa-actions-bar" style="margin-top: 15px;">
                <button type="button" class="btn btn-primary" onclick="startConvert()">
                    <span class="glyphicon glyphicon-play"></span> Start Conversion Run
                </button>
                <button type="button" class="btn btn-danger" onclick="cancelConvert()">
                    <span class="glyphicon glyphicon-stop"></span> Cancel Run
                </button>
                <span id="convert-status-text" style="color: #94a3b8; font-size: 13px; margin-left: 10px;">Status: Idle</span>
            </div>

            <div class="progress" style="background: #090e15; height: 18px; border-radius: 4px; display: none; margin-top: 20px;" id="convert-progress-bar-wrap">
                <div class="progress-bar progress-bar-striped active" role="progressbar" id="convert-progress-bar" style="width: 100%;">Converting...</div>
            </div>

            <div class="cwa-terminal" id="convert-terminal" style="display: none;">
[CWA] Convert Library Service initialized. Waiting for task execution...
            </div>
        </div>
    </div>

    <script src="/static/js/libs/jquery.min.js"></script>
    <script src="/static/js/libs/bootstrap.min.js"></script>
    <script>
    function startConvert() {
        document.getElementById('convert-progress-bar-wrap').style.display = 'block';
        document.getElementById('convert-terminal').style.display = 'block';
        document.getElementById('convert-status-text').innerText = 'Status: Running conversion...';
        fetch('/cwa-convert-library-start')
            .then(r => r.json())
            .then(d => {
                document.getElementById('convert-status-text').innerText = 'Status: ' + (d.message || 'Completed');
                document.getElementById('convert-terminal').innerText += '\n[CWA] ' + (d.message || 'Job finished');
            });
    }
    function cancelConvert() {
        fetch('/convert-library-cancel')
            .then(r => r.json())
            .then(d => {
                document.getElementById('convert-progress-bar-wrap').style.display = 'none';
                document.getElementById('convert-status-text').innerText = 'Status: Cancelled';
            });
    }
    </script>
</body>
</html>`

const cwaFixerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Calibre-Web Automated | EPUB Fixer Service</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link href="/static/css/libs/bootstrap.min.css" rel="stylesheet" media="screen">
    <link href="/static/css/style.css" rel="stylesheet" media="screen">
    <link href="/static/css/caliBlur.css" rel="stylesheet" media="screen">
    <link href="/static/css/caliBlur_override.css" rel="stylesheet" media="screen">
    <link href="/static/css/layout-refresh.css" rel="stylesheet" media="screen">
    <link href="/static/css/cwa.css?v=2026.5" rel="stylesheet" media="screen">
</head>
<body class="discover blur">
    <!-- Static navbar -->
    <div class="navbar navbar-default navbar-static-top" role="navigation">
        <div class="container-fluid">
            <div class="navbar-header">
                <button type="button" class="navbar-toggle" data-toggle="collapse" data-target=".navbar-collapse">
                    <span class="sr-only">Toggle Navigation</span>
                    <span class="icon-bar"></span>
                    <span class="icon-bar"></span>
                    <span class="icon-bar"></span>
                </button>
                <a class="navbar-brand" href="/"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#38bdf8" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 8px; flex-shrink: 0;"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path></svg>Calibre-Web Automated <span class="label label-warning" style="font-size: 10px; vertical-align: middle; margin-left: 5px;">CWA</span></a>
            </div>
            <div class="home-btn"><a class="home-btn-tooltip" href="/" data-toggle="tooltip" title="" data-placement="bottom" data-original-title="Home"></a></div>
            <div class="plexBack"><a href="/"></a></div>

            <form class="navbar-form navbar-left" role="search" action="/search" method="GET">
                <div class="form-group input-group input-group-sm">
                    <label for="query" class="sr-only">Search</label>
                    <input type="text" class="form-control" id="query" name="query" placeholder="Search Library">
                    <span class="input-group-btn">
                        <button type="submit" id="query_submit" class="btn btn-default"><span class="glyphicon glyphicon-search"></span></button>
                    </span>
                </div>
            </form>

            <div class="navbar-collapse collapse">
                <ul class="nav navbar-nav">
                    <li><a href="/advsearch" id="advanced_search"><span class="glyphicon glyphicon-search" style="padding-left: 10px; padding-right: 6px;"></span><span class="hidden-sm"> Advanced Search</span></a></li>
                </ul>
                <ul class="nav navbar-nav navbar-right" id="main-nav">
                    <li>
                        <form id="form-upload" class="navbar-form" action="/upload" method="post" enctype="multipart/form-data">
                            <div class="form-group">
                                <span class="btn btn-default btn-file"><span class="glyphicon glyphicon-upload"></span> Upload
                                    <input id="btn-upload" name="file" type="file" accept=".epub,.pdf,.mobi,.azw3,.cbr,.cbz" onchange="this.form.submit()">
                                </span>
                            </div>
                        </form>
                    </li>
                    <li class="top_tasks"><a id="top_tasks" href="/tasks"><span class="glyphicon glyphicon-tasks"></span> <span class="hidden-sm">Tasks</span></a></li>
                    <li><a id="top_admin" href="/admin"><span class="glyphicon glyphicon-dashboard"></span> <span class="hidden-sm">Admin</span></a></li>
                    <li class="cwa-switch-theme">
                        <a id="cwa-switch-theme" href="/cwa-switch-theme">
                            <span class="glyphicon glyphicon-sunglasses"></span>
                            <span class="hidden-sm">Switch Theme</span>
                        </a>
                    </li>
                    <li><a id="refresh-library" style="cursor: pointer;" href="javascript:void(0)" onclick="fetch('/cwa-library-refresh',{method:'POST'}).then(function(){location.reload();});"><span class="glyphicon glyphicon-refresh"></span> <span class="hidden-sm">Refresh Library</span></a></li>
                    <li class="dropdown">
                        <a href="#" class="dropdown-toggle profileDrop" data-toggle="dropdown" role="button" aria-haspopup="true" aria-expanded="false"><span class="glyphicon glyphicon-user"></span></a>
                        <ul class="dropdown-menu profileDropli">
                            <li><a id="top_user" href="/me"><span class="glyphicon glyphicon-user"></span> <span class="hidden-sm">Profile</span></a></li>
                            <li><a href="/cwa-settings"><span class="glyphicon glyphicon-cog"></span> CWA Settings</a></li>
                            <li><a href="/cwa-stats"><span class="glyphicon glyphicon-stats"></span> Deep Analytics</a></li>
                            <li><a href="/duplicates"><span class="glyphicon glyphicon-duplicate"></span> Duplicates</a></li>
                            <li><a href="/cwa-backup/list"><span class="glyphicon glyphicon-floppy-save"></span> Backups</a></li>
                            <li role="separator" class="divider"></li>
                            <li><a href="/logout"><span class="glyphicon glyphicon-log-out"></span> Logout</a></li>
                        </ul>
                    </li>
                </ul>
            </div>
        </div>
    </div>
    <div class="cwa-container">
        <div class="cwa-page-header">
            <div>
                <a href="/" class="btn btn-default cwa-back-btn" style="margin-bottom: 12px;">
                    <span class="glyphicon glyphicon-arrow-left"></span> Back to Library
                </a>
                <h1 class="cwa-page-title">
                    <span class="glyphicon glyphicon-wrench" style="color: #38bdf8;"></span>
                    Automatic EPUB Fixer Service
                </h1>
                <p class="cwa-page-subtitle">Kindle Send-to-Kindle compliance, XML sanitization, and encoding repair engine.</p>
            </div>
            <div>
                <span class="cwa-badge cwa-badge-warning">Send-to-Kindle 2.0</span>
            </div>
        </div>

        <div class="cwa-panel">
            <h2 class="cwa-panel-title">
                <span class="glyphicon glyphicon-info-sign"></span>
                Inspection &amp; Repair Rules
            </h2>
            <p class="cwa-panel-description">
                Eliminates Amazon Send-to-Kindle rejection emails and parsing errors:
            </p>
            <ul style="color: #94a3b8; font-size: 13px; line-height: 1.8; margin-bottom: 25px;">
                <li>Forces uncompressed <code>mimetype</code> store entry as the very first file in the ZIP archive.</li>
                <li>Normalizes malformed <code>&lt;dc:language&gt;</code> tags into valid BCP-47 identifiers.</li>
                <li>Sanitizes unclosed <code>&lt;img&gt;</code> and <code>&lt;br&gt;</code> tags causing XML parser aborts.</li>
                <li>Eliminates duplicate XML declarations and corrects broken NCX navigation hierarchies.</li>
            </ul>

            <div class="row">
                <div class="col-sm-6">
                    <div style="background: #0d151e; border: 1px solid rgba(255,255,255,0.06); border-radius: 6px; padding: 18px; margin-bottom: 20px;">
                        <h4 style="margin-top: 0; color: #38bdf8; font-size: 14px; font-weight: 600;">Repair Single Book</h4>
                        <div style="display: flex; gap: 8px; margin-top: 10px;">
                            <input type="number" id="single_book_id" class="form-control cwa-input" placeholder="Book ID (e.g. 1)" style="max-width: 180px;">
                            <button type="button" class="btn btn-info" onclick="fixSingleBook()">
                                <span class="glyphicon glyphicon-wrench"></span> Fix Book
                            </button>
                        </div>
                    </div>
                </div>

                <div class="col-sm-6">
                    <div style="background: #0d151e; border: 1px solid rgba(255,255,255,0.06); border-radius: 6px; padding: 18px; margin-bottom: 20px;">
                        <h4 style="margin-top: 0; color: #fbbf24; font-size: 14px; font-weight: 600;">Scan &amp; Fix Entire Library</h4>
                        <div class="cwa-actions-bar" style="margin-top: 10px;">
                            <button type="button" class="btn btn-warning" onclick="startFullScan()">
                                <span class="glyphicon glyphicon-play"></span> Run Library Fixer
                            </button>
                            <button type="button" class="btn btn-danger" onclick="cancelFixer()">
                                <span class="glyphicon glyphicon-stop"></span> Cancel
                            </button>
                        </div>
                    </div>
                </div>
            </div>

            <div style="margin-top: 10px;">
                <span id="fixer-status-text" style="color: #94a3b8; font-size: 13px;">Status: Ready</span>
            </div>

            <div class="progress" style="background: #090e15; height: 18px; border-radius: 4px; display: none; margin-top: 15px;" id="fixer-progress-bar-wrap">
                <div class="progress-bar progress-bar-warning progress-bar-striped active" role="progressbar" style="width: 100%;">Processing Library EPUBs...</div>
            </div>
        </div>
    </div>

    <script src="/static/js/libs/jquery.min.js"></script>
    <script src="/static/js/libs/bootstrap.min.js"></script>
    <script>
    function fixSingleBook() {
        const id = document.getElementById('single_book_id').value;
        if (!id) return alert('Please enter a book ID.');
        fetch('/cwa-epub-fixer-start?book_id=' + id)
            .then(r => r.json())
            .then(d => alert(d.message || 'Fixed!'));
    }
    function startFullScan() {
        document.getElementById('fixer-progress-bar-wrap').style.display = 'block';
        document.getElementById('fixer-status-text').innerText = 'Status: Fixing library EPUBs...';
        fetch('/cwa-epub-fixer-start')
            .then(r => r.json())
            .then(d => {
                document.getElementById('fixer-progress-bar-wrap').style.display = 'none';
                document.getElementById('fixer-status-text').innerText = 'Status: ' + (d.message || 'Completed');
                alert(d.message || 'Completed!');
            });
    }
    function cancelFixer() {
        fetch('/epub-fixer-cancel').then(() => {
            document.getElementById('fixer-progress-bar-wrap').style.display = 'none';
            document.getElementById('fixer-status-text').innerText = 'Status: Cancelled';
        });
    }
    </script>
</body>
</html>`

const cwaBackupHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Calibre-Web Automated | Backups</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link href="/static/css/libs/bootstrap.min.css" rel="stylesheet" media="screen">
    <link href="/static/css/style.css" rel="stylesheet" media="screen">
    <link href="/static/css/caliBlur.css" rel="stylesheet" media="screen">
    <link href="/static/css/caliBlur_override.css" rel="stylesheet" media="screen">
    <link href="/static/css/layout-refresh.css" rel="stylesheet" media="screen">
    <link href="/static/css/cwa.css?v=2026.5" rel="stylesheet" media="screen">
</head>
<body class="discover blur">
    <!-- Static navbar -->
    <div class="navbar navbar-default navbar-static-top" role="navigation">
        <div class="container-fluid">
            <div class="navbar-header">
                <button type="button" class="navbar-toggle" data-toggle="collapse" data-target=".navbar-collapse">
                    <span class="sr-only">Toggle Navigation</span>
                    <span class="icon-bar"></span>
                    <span class="icon-bar"></span>
                    <span class="icon-bar"></span>
                </button>
                <a class="navbar-brand" href="/"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#38bdf8" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 8px; flex-shrink: 0;"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path></svg>Calibre-Web Automated <span class="label label-warning" style="font-size: 10px; vertical-align: middle; margin-left: 5px;">CWA</span></a>
            </div>
            <div class="home-btn"><a class="home-btn-tooltip" href="/" data-toggle="tooltip" title="" data-placement="bottom" data-original-title="Home"></a></div>
            <div class="plexBack"><a href="/"></a></div>

            <form class="navbar-form navbar-left" role="search" action="/search" method="GET">
                <div class="form-group input-group input-group-sm">
                    <label for="query" class="sr-only">Search</label>
                    <input type="text" class="form-control" id="query" name="query" placeholder="Search Library">
                    <span class="input-group-btn">
                        <button type="submit" id="query_submit" class="btn btn-default"><span class="glyphicon glyphicon-search"></span></button>
                    </span>
                </div>
            </form>

            <div class="navbar-collapse collapse">
                <ul class="nav navbar-nav">
                    <li><a href="/advsearch" id="advanced_search"><span class="glyphicon glyphicon-search" style="padding-left: 10px; padding-right: 6px;"></span><span class="hidden-sm"> Advanced Search</span></a></li>
                </ul>
                <ul class="nav navbar-nav navbar-right" id="main-nav">
                    <li>
                        <form id="form-upload" class="navbar-form" action="/upload" method="post" enctype="multipart/form-data">
                            <div class="form-group">
                                <span class="btn btn-default btn-file"><span class="glyphicon glyphicon-upload"></span> Upload
                                    <input id="btn-upload" name="file" type="file" accept=".epub,.pdf,.mobi,.azw3,.cbr,.cbz" onchange="this.form.submit()">
                                </span>
                            </div>
                        </form>
                    </li>
                    <li class="top_tasks"><a id="top_tasks" href="/tasks"><span class="glyphicon glyphicon-tasks"></span> <span class="hidden-sm">Tasks</span></a></li>
                    <li><a id="top_admin" href="/admin"><span class="glyphicon glyphicon-dashboard"></span> <span class="hidden-sm">Admin</span></a></li>
                    <li class="cwa-switch-theme">
                        <a id="cwa-switch-theme" href="/cwa-switch-theme">
                            <span class="glyphicon glyphicon-sunglasses"></span>
                            <span class="hidden-sm">Switch Theme</span>
                        </a>
                    </li>
                    <li><a id="refresh-library" style="cursor: pointer;" href="javascript:void(0)" onclick="fetch('/cwa-library-refresh',{method:'POST'}).then(function(){location.reload();});"><span class="glyphicon glyphicon-refresh"></span> <span class="hidden-sm">Refresh Library</span></a></li>
                    <li class="dropdown">
                        <a href="#" class="dropdown-toggle profileDrop" data-toggle="dropdown" role="button" aria-haspopup="true" aria-expanded="false"><span class="glyphicon glyphicon-user"></span></a>
                        <ul class="dropdown-menu profileDropli">
                            <li><a id="top_user" href="/me"><span class="glyphicon glyphicon-user"></span> <span class="hidden-sm">Profile</span></a></li>
                            <li><a href="/cwa-settings"><span class="glyphicon glyphicon-cog"></span> CWA Settings</a></li>
                            <li><a href="/cwa-stats"><span class="glyphicon glyphicon-stats"></span> Deep Analytics</a></li>
                            <li><a href="/duplicates"><span class="glyphicon glyphicon-duplicate"></span> Duplicates</a></li>
                            <li><a href="/cwa-backup/list"><span class="glyphicon glyphicon-floppy-save"></span> Backups</a></li>
                            <li role="separator" class="divider"></li>
                            <li><a href="/logout"><span class="glyphicon glyphicon-log-out"></span> Logout</a></li>
                        </ul>
                    </li>
                </ul>
            </div>
        </div>
    </div>
    <div class="cwa-container">
        <div class="cwa-page-header">
            <div>
                <a href="/" class="btn btn-default cwa-back-btn" style="margin-bottom: 12px;">
                    <span class="glyphicon glyphicon-arrow-left"></span> Back to Library
                </a>
                <h1 class="cwa-page-title">
                    <span class="glyphicon glyphicon-floppy-save" style="color: #38bdf8;"></span>
                    Automated Back Up Service
                </h1>
                <p class="cwa-page-subtitle">Compressed snapshot management for Calibre metadata databases and settings.</p>
            </div>
            <div>
                <button type="button" class="btn btn-success" onclick="createBackup()">
                    <span class="glyphicon glyphicon-plus"></span> Create Compressed Snapshot (.zip)
                </button>
            </div>
        </div>

        <div class="cwa-panel">
            <h2 class="cwa-panel-title">
                <span class="glyphicon glyphicon-list"></span>
                Available Snapshots
            </h2>
            <p class="cwa-panel-description">
                Snapshots include <code>metadata.db</code>, <code>app.db</code>, and active SQLite WAL journals in Deflate compressed archive format.
            </p>

            <table class="table cwa-table">
                <thead>
                    <tr>
                        <th>Archive Filename</th>
                        <th>Size</th>
                        <th>Timestamp</th>
                        <th class="text-right">Actions</th>
                    </tr>
                </thead>
                <tbody id="backups-tbody">
                    {% for b in backups %}
                    <tr>
                        <td><code>{{ b.Filename }}</code></td>
                        <td>{{ b.SizeHuman }}</td>
                        <td>{{ b.CreatedAt|date:"2006-01-02 15:04:05" }}</td>
                        <td class="text-right">
                            <a href="/cwa-backup/download/{{ b.Filename }}" class="btn btn-xs btn-primary">
                                <span class="glyphicon glyphicon-download-alt"></span> Download
                            </a>
                            <button type="button" class="btn btn-xs btn-danger" onclick="deleteBackup('{{ b.Filename }}')">
                                <span class="glyphicon glyphicon-trash"></span> Delete
                            </button>
                        </td>
                    </tr>
                    {% empty %}
                    <tr>
                        <td colspan="4" class="text-center" style="padding: 30px; color: #8194a7;">
                            No snapshots have been generated yet. Press "Create Compressed Snapshot" above to generate one.
                        </td>
                    </tr>
                    {% endfor %}
                </tbody>
            </table>
        </div>
    </div>

    <script src="/static/js/libs/jquery.min.js"></script>
    <script src="/static/js/libs/bootstrap.min.js"></script>
    <script>
    function createBackup() {
        fetch('/cwa-backup/create', {method: 'POST'})
            .then(r => r.json())
            .then(d => {
                alert(d.message || 'Snapshot generated successfully.');
                location.reload();
            });
    }
    function deleteBackup(filename) {
        if (!confirm('Permanently delete backup snapshot ' + filename + '?')) return;
        fetch('/cwa-backup/delete/' + filename, {method: 'POST'})
            .then(r => r.json())
            .then(d => {
                location.reload();
            });
    }
    </script>
</body>
</html>`
