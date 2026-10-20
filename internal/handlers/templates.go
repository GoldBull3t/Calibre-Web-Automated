package handlers

const gridHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Calibre-Web Automated | {{ title }}</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1">
    <link rel="apple-touch-icon" sizes="140x140" href="/static/favicon.ico">
    <link rel="shortcut icon" href="/static/favicon.ico">
    <link href="/static/css/libs/bootstrap.min.css" rel="stylesheet" media="screen">
    <link href="/static/css/style.css" rel="stylesheet" media="screen">
    <link href="/static/css/upload.css" rel="stylesheet" media="screen">
    <link href="/static/css/caliBlur.css" rel="stylesheet" media="screen">
    <link href="/static/css/caliBlur_override.css" rel="stylesheet" media="screen">
    <link href="/static/css/layout-refresh.css" rel="stylesheet" media="screen">
    <link href="/static/css/cwa.css?v=2026.5" rel="stylesheet" media="screen">
</head>
<body class="discover load-more blur allow-mobile-blur" data-text="Home" data-textback="Back">
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
                    <input type="text" class="form-control" id="query" name="query" placeholder="Search Library" value="{{ query }}">
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

    <div id="loader" hidden="true">
        <center>
            <h3>Uploading...</h3>
            <span>Please do not refresh the page</span>.
        </center>
    </div>

    <div class="container-fluid">
        <div class="row-fluid">
            <div class="col-sm-2">
                <nav class="navigation">
                    <ul class="list-unstyled" id="scnd-nav" intent in-standard-append="nav.navigation" in-mobile-after="#main-nav" in-mobile-class="nav navbar-nav">
                        <li class="nav-head hidden-xs">Browse</li>
                        <li id="nav_new" class="{% if activeNav == "new" %}active{% endif %}">
                            <a href="/"><span class="glyphicon glyphicon-book"></span> Books</a>
                        </li>
                        <li id="nav_hot" class="{% if activeNav == "hot" %}active{% endif %}">
                            <a href="/hot"><span class="glyphicon glyphicon-fire"></span> Hot Books</a>
                        </li>
                        <li id="nav_rand" class="{% if activeNav == "rand" %}active{% endif %}">
                            <a href="/discover"><span class="glyphicon glyphicon-random"></span> Discover</a>
                        </li>
                        <li id="nav_author" class="{% if activeNav == "author" %}active{% endif %}">
                            <a href="/authors"><span class="glyphicon glyphicon-user"></span> Authors</a>
                        </li>
                        <li id="nav_series" class="{% if activeNav == "series" %}active{% endif %}">
                            <a href="/series"><span class="glyphicon glyphicon-list"></span> Series</a>
                        </li>
                        <li id="nav_category" class="{% if activeNav == "category" %}active{% endif %}">
                            <a href="/category"><span class="glyphicon glyphicon-tag"></span> Categories</a>
                        </li>
                        <li id="nav_format" class="{% if activeNav == "format" %}active{% endif %}">
                            <a href="/formats"><span class="glyphicon glyphicon-file"></span> Formats</a>
                        </li>
                        <li id="nav_rated" class="{% if activeNav == "ratings" %}active{% endif %}">
                            <a href="/ratings"><span class="glyphicon glyphicon-star"></span> Top Rated</a>
                        </li>
                        <li id="nav_publisher" class="{% if activeNav == "publisher" %}active{% endif %}">
                            <a href="/publishers"><span class="glyphicon glyphicon-briefcase"></span> Publishers</a>
                        </li>
                        <li id="nav_language" class="{% if activeNav == "language" %}active{% endif %}">
                            <a href="/languages"><span class="glyphicon glyphicon-globe"></span> Languages</a>
                        </li>
                        <li id="nav_read" class="{% if activeNav == "read" %}active{% endif %}">
                            <a href="/readbooks"><span class="glyphicon glyphicon-ok"></span> Read Books</a>
                        </li>
                        <li id="nav_unread" class="{% if activeNav == "unread" %}active{% endif %}">
                            <a href="/unreadbooks"><span class="glyphicon glyphicon-eye-open"></span> Unread Books</a>
                        </li>

                        <li class="nav-head hidden-xs" style="margin-top: 15px;">Magic Shelves</li>
                        <li><a href="/magicshelf/1"><span class="glyphicon glyphicon-time"></span> Recently Added</a></li>
                        <li><a href="/magicshelf/2"><span class="glyphicon glyphicon-star"></span> Highly Rated</a></li>
                        <li><a href="/magicshelf/3"><span class="glyphicon glyphicon-bookmark"></span> Currently Reading</a></li>
                        <li><a href="/magicshelf/4"><span class="glyphicon glyphicon-picture"></span> Books Without Covers</a></li>

                        <li class="nav-head hidden-xs" style="margin-top: 15px;">CWA Automation</li>
                        <li id="nav_cwa_settings"><a href="/cwa-settings"><span class="glyphicon glyphicon-cog"></span> CWA Settings</a></li>
                        <li id="nav_duplicates" class="{% if activeNav == "duplicates" %}active{% endif %}"><a href="/duplicates"><span class="glyphicon glyphicon-duplicate"></span> Duplicates Manager</a></li>
                        <li id="nav_cwa_convert"><a href="/cwa-convert-library-overview"><span class="glyphicon glyphicon-refresh"></span> Convert Library</a></li>
                        <li id="nav_cwa_fixer"><a href="/cwa-epub-fixer-overview"><span class="glyphicon glyphicon-wrench"></span> EPUB Fixer Service</a></li>
                        <li id="nav_cwa_stats"><a href="/cwa-stats"><span class="glyphicon glyphicon-stats"></span> Deep Analytics</a></li>
                        <li id="nav_cwa_backup"><a href="/cwa-backup/list"><span class="glyphicon glyphicon-floppy-save"></span> Backups</a></li>
                        <li id="nav_opds"><a href="/opds"><span class="glyphicon glyphicon-transfer"></span> OPDS Feed</a></li>
                    </ul>
                </nav>
            </div>

            <div class="col-sm-10">
                <div class="shelf-actions">
                    <h2 style="margin: 0; font-size: 24px; color: #fff;">{{ title }} <span style="font-size: 16px; color: #888;">({{ result.TotalItems }} books)</span></h2>
                    <div class="filterheader">
                        <div class="btn-group">
                            <button type="button" class="btn btn-default dropdown-toggle" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
                                Sort: {% if sort == "abc" %}Title A-Z{% elif sort == "zyx" %}Title Z-A{% elif sort == "old" %}Oldest{% elif sort == "authaz" %}Author A-Z{% elif sort == "authza" %}Author Z-A{% elif sort == "pubnew" %}Publication Date{% elif sort == "random" %}Random{% else %}Newest{% endif %} <span class="caret"></span>
                            </button>
                            <ul class="dropdown-menu dropdown-menu-right">
                                <li><a href="?sort=new{% if query %}&query={{ query|urlencode }}{% endif %}">Newest</a></li>
                                <li><a href="?sort=old{% if query %}&query={{ query|urlencode }}{% endif %}">Oldest</a></li>
                                <li><a href="?sort=abc{% if query %}&query={{ query|urlencode }}{% endif %}">Title (A-Z)</a></li>
                                <li><a href="?sort=zyx{% if query %}&query={{ query|urlencode }}{% endif %}">Title (Z-A)</a></li>
                                <li><a href="?sort=authaz{% if query %}&query={{ query|urlencode }}{% endif %}">Author (A-Z)</a></li>
                                <li><a href="?sort=authza{% if query %}&query={{ query|urlencode }}{% endif %}">Author (Z-A)</a></li>
                                <li><a href="?sort=pubnew{% if query %}&query={{ query|urlencode }}{% endif %}">Publication Date</a></li>
                                <li><a href="?sort=random{% if query %}&query={{ query|urlencode }}{% endif %}">Random</a></li>
                            </ul>
                        </div>
                    </div>
                </div>

                <div class="row">
                    {% for b in result.Items %}
                    <div class="col-xs-6 col-sm-4 col-md-3 col-lg-2" style="margin-bottom: 25px;">
                        <div class="book" style="text-align: center;">
                            <div class="cover">
                                <a href="/book/{{ b.ID }}" data-toggle="modal" data-target="#bookDetailsModal" data-remote="/book/{{ b.ID }}?modal=1">
                                    <span class="img">
                                        <img src="/cover/{{ b.ID }}" alt="{{ b.Title }}" style="width: 100%; height: 230px; object-fit: cover; border-radius: 4px; box-shadow: 0 4px 8px rgba(0,0,0,0.4);" loading="lazy">
                                    </span>
                                </a>
                            </div>
                            <div class="meta" style="margin-top: 8px;">
                                <p class="title" style="margin: 0; font-weight: bold; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">
                                    <a href="/book/{{ b.ID }}" title="{{ b.Title }}" style="color: #eee;">{{ b.Title }}</a>
                                </p>
                                <p class="author" style="margin: 0; font-size: 12px; color: #aaa; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">
                                    {% for a in b.Authors %}
                                        <a href="/author/stored/{{ a }}" style="color: #aaa;">{{ a }}</a>{% if not forloop.Last %}, {% endif %}
                                    {% endfor %}
                                </p>
                                {% if b.Rating > 0 %}
                                <div class="rating" style="color: #f0ad4e; font-size: 11px; margin-top: 2px;">
                                    {% if b.Rating >= 2 %}★{% endif %}{% if b.Rating >= 4 %}★{% endif %}{% if b.Rating >= 6 %}★{% endif %}{% if b.Rating >= 8 %}★{% endif %}{% if b.Rating >= 10 %}★{% endif %}
                                </div>
                                {% endif %}
                            </div>
                        </div>
                    </div>
                    {% empty %}
                    <div class="col-xs-12">
                        <div class="cwa-empty-state">
                            <div class="cwa-empty-icon">
                                <span class="glyphicon glyphicon-book"></span>
                            </div>
                            <h3 class="cwa-empty-title">Your Library is Empty</h3>
                            <p class="cwa-empty-desc">
                                Add your favorite books or drop files into the automated ingest folder. Supported formats include EPUB, PDF, MOBI, AZW3, CBR and CBZ.
                            </p>
                            <div class="cwa-empty-actions">
                                <label class="btn btn-primary" style="cursor: pointer; margin-bottom: 0;">
                                    <span class="glyphicon glyphicon-upload"></span> Upload First Book
                                    <input type="file" accept=".epub,.pdf,.mobi,.azw3,.cbr,.cbz" onchange="document.getElementById('btn-upload').files = this.files; document.getElementById('form-upload').submit();" style="display: none;">
                                </label>
                                <a href="/cwa-settings" class="btn btn-default">
                                    <span class="glyphicon glyphicon-cog"></span> CWA Settings
                                </a>
                            </div>
                        </div>
                    </div>
                    {% endfor %}
                </div>

                {% if result.TotalPages > 1 %}
                <div class="text-center">
                    <ul class="pagination">
                        {% if result.Page > 1 %}
                        <li><a href="?page={{ result.Page - 1 }}{% if sort %}&sort={{ sort }}{% endif %}{% if query %}&query={{ query|urlencode }}{% endif %}">&laquo;</a></li>
                        {% endif %}
                        <li class="active"><span>Page {{ result.Page }} of {{ result.TotalPages }}</span></li>
                        {% if result.Page < result.TotalPages %}
                        <li><a href="?page={{ result.Page + 1 }}{% if sort %}&sort={{ sort }}{% endif %}{% if query %}&query={{ query|urlencode }}{% endif %}">&raquo;</a></li>
                        {% endif %}
                    </ul>
                </div>
                {% endif %}
            </div>
        </div>
    </div>

    <!-- Modal de Detalhes -->
    <div class="modal fade" id="bookDetailsModal" tabindex="-1" role="dialog" aria-labelledby="bookDetailsModalLabel">
        <div class="modal-dialog modal-lg" role="document">
            <div class="modal-content" style="background: #18222d; color: #fff; border: 1px solid rgba(255,255,255,0.1); border-radius: 12px;">
                <div class="modal-header" style="border-bottom: 1px solid rgba(255,255,255,0.1);">
                    <button type="button" class="close" data-dismiss="modal" aria-label="Close" style="color: #fff; opacity: 0.8;"><span aria-hidden="true">&times;</span></button>
                    <h4 class="modal-title" id="bookDetailsModalLabel">Book Details</h4>
                </div>
                <div class="modal-body" style="padding: 20px;">
                    <div class="text-center"><span class="glyphicon glyphicon-refresh glyphicon-refresh-animate"></span> Loading...</div>
                </div>
                <div class="modal-footer" style="border-top: 1px solid rgba(255,255,255,0.1);">
                    <button type="button" class="btn btn-default" data-dismiss="modal">Close</button>
                </div>
            </div>
        </div>
    </div>

    <script src="/static/js/libs/jquery.min.js"></script>
    <script src="/static/js/libs/bootstrap.min.js"></script>
    <script>window.bootstrap = window.bootstrap || {};</script>
    <script src="/static/js/libs/underscore-umd-min.js"></script>
    <script src="/static/js/libs/intention.min.js"></script>
    <script src="/static/js/libs/context.min.js"></script>
    <script src="/static/js/libs/plugins.js"></script>
    <script src="/static/js/libs/jquery.form.min.js"></script>
    <script src="/static/js/uploadprogress.js"></script>
    <script src="/static/js/main.js"></script>
    <script src="/static/js/libs/jquery.visible.min.js"></script>
    <script src="/static/js/libs/compromise.min.js"></script>
    <script src="/static/js/libs/readmore.min.js"></script>
    <script>
        window.scriptRoot = "";
        window.cwaUserData = { canEditBooks: true };
    </script>
    <script src="/static/js/caliBlur.js"></script>
    <script>
    $('#bookDetailsModal').on('show.bs.modal', function (e) {
        var button = $(e.relatedTarget);
        var url = button.data('remote');
        var modal = $(this);
        modal.find('.modal-body').load(url);
    });
    </script>
</body>
</html>`

const modalHTML = `<div class="book-detail-card" style="width: 100%; background: transparent; border: none; padding: 0; box-shadow: none;">
    <div class="book-detail-main" style="display: flex; gap: 28px; align-items: flex-start; flex-wrap: wrap;">
        <div class="book-detail-cover" style="flex: 0 0 190px; max-width: 190px;">
            <img src="/cover/{{ book.ID }}" alt="{{ book.Title }}" style="width: 100%; height: auto; border-radius: 8px; box-shadow: 0 10px 25px rgba(0,0,0,0.5);">
        </div>
        <div class="book-detail-meta" style="flex: 1 1 300px; display: flex; flex-direction: column; gap: 8px;">
            <h2 id="title" style="margin: 0; font-size: 24px; font-weight: 700; color: #fff; letter-spacing: -0.3px;">{{ book.Title }}</h2>
            <p class="author" style="margin: 0; font-size: 15px;">
                {% for a in book.Authors %}
                    <a href="/author/stored/{{ a }}" style="color: #38bdf8; font-weight: 500;">{{ a }}</a>{% if not forloop.Last %}, {% endif %}
                {% endfor %}
            </p>
            {% if book.Series %}
            <p style="margin: 0; font-size: 13px; color: #94a3b8;">
                Series: <a href="/series/stored/{{ book.Series }}" style="color: #cbd5e1;">{{ book.Series }}</a> <span class="cwa-badge cwa-badge-info" style="font-size: 10px; padding: 2px 6px;">#{{ book.SeriesIndex }}</span>
            </p>
            {% endif %}
            <div class="book-metadata" style="display: flex; flex-wrap: wrap; gap: 6px; margin: 8px 0;">
                {% if book.Rating > 0 %}
                <span class="cwa-badge cwa-badge-warning">★ {{ book.Rating / 2 }} Stars</span>
                {% endif %}
                {% for t in book.Tags %}
                <span class="cwa-badge cwa-badge-info"><a href="/category/stored/{{ t }}" style="color: inherit; text-decoration: none;">{{ t }}</a></span>
                {% endfor %}
                {% for p in book.Publishers %}
                <span class="cwa-badge" style="background: rgba(255,255,255,0.06); color: #cbd5e1; border: 1px solid rgba(255,255,255,0.1);">{{ p }}</span>
                {% endfor %}
            </div>

            <div class="book-detail-actions" style="margin-top: 15px; display: flex; flex-wrap: wrap; gap: 8px;">
                <a href="/read/{{ book.ID }}" target="_blank" class="btn btn-primary btn-sm"><span class="glyphicon glyphicon-book"></span> Read in Browser</a>
                <div class="btn-group">
                    <button type="button" class="btn btn-default btn-sm dropdown-toggle" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
                        <span class="glyphicon glyphicon-download-alt"></span> Download <span class="caret"></span>
                    </button>
                    <ul class="dropdown-menu">
                        {% for f in book.Formats %}
                        <li><a href="/download/{{ book.ID }}/{{ f.Format }}">{{ f.Format }} ({{ f.Size|filesizeformat }})</a></li>
                        {% endfor %}
                    </ul>
                </div>
                <a href="/admin/book/{{ book.ID }}" class="btn btn-default btn-sm"><span class="glyphicon glyphicon-pencil"></span> Edit Metadata</a>
                <button type="button" class="btn btn-default btn-sm" onclick="fetch('/ajax/toggleread/{{ book.ID }}',{method:'POST'}).then(function(){alert('Status updated');});"><span class="glyphicon glyphicon-check"></span> Toggle Read</button>
                <button type="button" class="btn btn-danger btn-sm" onclick="if(confirm('Delete book?')){fetch('/delete/{{ book.ID }}',{method:'POST'}).then(function(){location.reload();});}"><span class="glyphicon glyphicon-trash"></span> Delete</button>
            </div>
        </div>
    </div>
    {% if book.Comment %}
    <div class="book-detail-synopsis" style="margin-top: 24px; border-top: 1px solid rgba(255,255,255,0.08); padding-top: 16px;">
        <h4 style="color: #64748b; font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.08em; margin-top: 0; margin-bottom: 8px;">Description</h4>
        <div style="line-height: 1.6; color: #cbd5e1; font-size: 13.5px; max-height: 220px; overflow-y: auto;">
            {{ safeDesc }}
        </div>
    </div>
    {% endif %}
</div>`

const detailHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Calibre-Web | {{ title }}</title>
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
    <div id="loader" hidden="true"></div>
    <div class="navbar navbar-default navbar-static-top" role="navigation">
        <div class="container-fluid">
            <div class="navbar-header">
                <a class="navbar-brand" href="/"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#38bdf8" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 8px; flex-shrink: 0;"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path></svg>Calibre-Web Automated <span class="label label-warning" style="font-size: 10px; vertical-align: middle; margin-left: 5px;">CWA</span></a>
            </div>
            <form class="navbar-form navbar-left" role="search" action="/search" method="GET">
                <div class="form-group input-group input-group-sm">
                    <input type="text" class="form-control" id="query" name="query" placeholder="Search Library">
                </div>
            </form>
            <div class="navbar-collapse collapse">
                <ul class="nav navbar-nav navbar-right" id="main-nav">
                    <li><a href="/"><span class="glyphicon glyphicon-book"></span> Library</a></li>
                    <li><a href="/cwa-settings"><span class="glyphicon glyphicon-cog"></span> CWA Settings</a></li>
                    <li><a href="/cwa-stats"><span class="glyphicon glyphicon-stats"></span> Deep Analytics</a></li>
                </ul>
            </div>
        </div>
    </div>

    <div class="container" style="max-width: 900px; margin-top: 76px; margin-bottom: 60px;">
        <div style="margin-bottom: 18px;">
            <a href="/" class="btn btn-default cwa-back-btn"><span class="glyphicon glyphicon-arrow-left"></span> Back to Library</a>
        </div>
        <div class="panel panel-default" style="background: rgba(18, 25, 38, 0.85); border-color: rgba(255,255,255,0.08); padding: 32px; border-radius: 14px; box-shadow: 0 16px 40px rgba(0,0,0,0.4); backdrop-filter: blur(16px);">
            ` + modalHTML + `
        </div>
    </div>
    <script src="/static/js/libs/jquery.min.js"></script>
    <script src="/static/js/libs/bootstrap.min.js"></script>
</body>
</html>`

const entitiesHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Calibre-Web Automated | {{ title }}</title>
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
    <div id="loader" hidden="true"></div>
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
                    <li class="top_tasks"><a id="top_tasks" href="/tasks"><span class="glyphicon glyphicon-tasks"></span> <span class="hidden-sm">Tasks</span></a></li>
                    <li><a id="top_admin" href="/admin"><span class="glyphicon glyphicon-dashboard"></span> <span class="hidden-sm">Admin</span></a></li>
                    <li class="cwa-switch-theme"><a id="cwa-switch-theme" href="/cwa-switch-theme"><span class="glyphicon glyphicon-sunglasses"></span> <span class="hidden-sm">Switch Theme</span></a></li>
                    <li><a id="refresh-library" style="cursor: pointer;" href="javascript:void(0)" onclick="fetch('/cwa-library-refresh',{method:'POST'}).then(function(){location.reload();});"><span class="glyphicon glyphicon-refresh"></span> <span class="hidden-sm">Refresh</span></a></li>
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

    <div class="container-fluid">
        <div class="row-fluid">
            <div class="col-sm-2">
                <nav class="navigation">
                    <ul class="list-unstyled" id="scnd-nav">
                        <li class="nav-head hidden-xs">Browse</li>
                        <li id="nav_new"><a href="/"><span class="glyphicon glyphicon-book"></span> Books</a></li>
                        <li id="nav_hot"><a href="/hot"><span class="glyphicon glyphicon-fire"></span> Hot Books</a></li>
                        <li id="nav_rand"><a href="/discover"><span class="glyphicon glyphicon-random"></span> Discover</a></li>
                        <li id="nav_author" class="{% if activeNav == "author" %}active{% endif %}"><a href="/authors"><span class="glyphicon glyphicon-user"></span> Authors</a></li>
                        <li id="nav_series" class="{% if activeNav == "series" %}active{% endif %}"><a href="/series"><span class="glyphicon glyphicon-list"></span> Series</a></li>
                        <li id="nav_category" class="{% if activeNav == "category" %}active{% endif %}"><a href="/category"><span class="glyphicon glyphicon-tag"></span> Categories</a></li>
                        <li id="nav_format" class="{% if activeNav == "format" %}active{% endif %}"><a href="/formats"><span class="glyphicon glyphicon-file"></span> Formats</a></li>
                        <li id="nav_rated" class="{% if activeNav == "ratings" %}active{% endif %}"><a href="/ratings"><span class="glyphicon glyphicon-star"></span> Top Rated</a></li>
                        <li id="nav_publisher" class="{% if activeNav == "publisher" %}active{% endif %}"><a href="/publishers"><span class="glyphicon glyphicon-briefcase"></span> Publishers</a></li>
                        <li id="nav_language" class="{% if activeNav == "language" %}active{% endif %}"><a href="/languages"><span class="glyphicon glyphicon-globe"></span> Languages</a></li>
                        <li id="nav_read"><a href="/readbooks"><span class="glyphicon glyphicon-ok"></span> Read Books</a></li>
                        <li id="nav_unread"><a href="/unreadbooks"><span class="glyphicon glyphicon-eye-open"></span> Unread Books</a></li>

                        <li class="nav-head hidden-xs" style="margin-top: 15px;">Magic Shelves</li>
                        <li><a href="/magicshelf/1"><span class="glyphicon glyphicon-time"></span> Recently Added</a></li>
                        <li><a href="/magicshelf/2"><span class="glyphicon glyphicon-star"></span> Highly Rated</a></li>
                        <li><a href="/magicshelf/3"><span class="glyphicon glyphicon-bookmark"></span> Currently Reading</a></li>
                        <li><a href="/magicshelf/4"><span class="glyphicon glyphicon-picture"></span> Books Without Covers</a></li>

                        <li class="nav-head hidden-xs" style="margin-top: 15px;">CWA Automation</li>
                        <li id="nav_cwa_settings"><a href="/cwa-settings"><span class="glyphicon glyphicon-cog"></span> CWA Settings</a></li>
                        <li id="nav_duplicates"><a href="/duplicates"><span class="glyphicon glyphicon-duplicate"></span> Duplicates Manager</a></li>
                        <li id="nav_cwa_convert"><a href="/cwa-convert-library-overview"><span class="glyphicon glyphicon-refresh"></span> Convert Library</a></li>
                        <li id="nav_cwa_fixer"><a href="/cwa-epub-fixer-overview"><span class="glyphicon glyphicon-wrench"></span> EPUB Fixer Service</a></li>
                        <li id="nav_cwa_stats"><a href="/cwa-stats"><span class="glyphicon glyphicon-stats"></span> Deep Analytics</a></li>
                        <li id="nav_cwa_backup"><a href="/cwa-backup/list"><span class="glyphicon glyphicon-floppy-save"></span> Backups</a></li>
                        <li id="nav_opds"><a href="/opds"><span class="glyphicon glyphicon-transfer"></span> OPDS Feed</a></li>
                    </ul>
                </nav>
            </div>

            <div class="col-sm-10">
                <div class="shelf-actions">
                    <h2>{{ title }} <span>({{ totalItems }} items)</span></h2>
                </div>

                <div class="row">
                    {% for it in items %}
                    <div class="col-xs-12 col-sm-6 col-md-4 col-lg-3" style="margin-bottom: 12px;">
                        <a href="{{ it.URL }}" class="cwa-entity-card">
                            <span class="cwa-entity-name">
                                <span class="glyphicon {{ it.Icon }}" style="margin-right: 8px; color: #38bdf8;"></span>{{ it.Name }}
                            </span>
                            <span class="cwa-badge cwa-badge-info">{{ it.Count }}</span>
                        </a>
                    </div>
                    {% empty %}
                    <div class="col-xs-12 text-center" style="padding: 50px 0;">
                        <p style="font-size: 16px; color: #64748b;">No entries available.</p>
                    </div>
                    {% endfor %}
                </div>

                {% if totalPages > 1 %}
                <div class="text-center" style="margin-top: 30px;">
                    <ul class="pagination">
                        {% if page > 1 %}
                        <li><a href="{{ basePath }}?page={{ page - 1 }}">&laquo;</a></li>
                        {% endif %}
                        <li class="active"><span>Page {{ page }} of {{ totalPages }}</span></li>
                        {% if page < totalPages %}
                        <li><a href="{{ basePath }}?page={{ page + 1 }}">&raquo;</a></li>
                        {% endif %}
                    </ul>
                </div>
                {% endif %}
            </div>
        </div>
    </div>
    <script src="/static/js/libs/jquery.min.js"></script>
    <script src="/static/js/libs/bootstrap.min.js"></script>
</body>
</html>`

const searchFormHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Calibre-Web | Advanced Search</title>
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
    <div id="loader" hidden="true"></div>
    <div class="navbar navbar-default navbar-static-top" role="navigation">
        <div class="container-fluid">
            <div class="navbar-header">
                <a class="navbar-brand" href="/"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#38bdf8" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 8px; flex-shrink: 0;"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path></svg>Calibre-Web Automated <span class="label label-warning" style="font-size: 10px; vertical-align: middle; margin-left: 5px;">CWA</span></a>
            </div>
            <form class="navbar-form navbar-left" role="search" action="/search" method="GET">
                <div class="form-group input-group input-group-sm">
                    <input type="text" class="form-control" id="query_top" name="query" placeholder="Search Library">
                </div>
            </form>
            <div class="navbar-collapse collapse">
                <ul class="nav navbar-nav navbar-right" id="main-nav">
                    <li><a href="/"><span class="glyphicon glyphicon-book"></span> Library</a></li>
                    <li><a href="/cwa-settings"><span class="glyphicon glyphicon-cog"></span> CWA Settings</a></li>
                    <li><a href="/cwa-stats"><span class="glyphicon glyphicon-stats"></span> Deep Analytics</a></li>
                </ul>
            </div>
        </div>
    </div>

    <div class="container" style="max-width: 800px; margin-top: 76px; margin-bottom: 60px;">
        <div style="margin-bottom: 18px;">
            <a href="/" class="btn btn-default cwa-back-btn"><span class="glyphicon glyphicon-arrow-left"></span> Back to Library</a>
        </div>
        <div class="panel panel-default" style="background: rgba(18, 25, 38, 0.85); border-color: rgba(255,255,255,0.08); border-radius: 14px; padding: 32px; color: #fff; box-shadow: 0 16px 40px rgba(0,0,0,0.4); backdrop-filter: blur(16px);">
            <h2 style="margin-top: 0; margin-bottom: 24px; border-bottom: 1px solid rgba(255,255,255,0.08); padding-bottom: 14px; font-size: 22px; font-weight: 700; letter-spacing: -0.3px;">Advanced Search</h2>
            <form action="/search" method="GET" role="form">
                <div class="form-group">
                    <label for="query" style="color: #cbd5e1; font-weight: 600;">Keyword / Title</label>
                    <input type="text" class="form-control" name="query" id="query" placeholder="Enter keywords or title..." style="background: rgba(255,255,255,0.05); color: #fff; border-color: rgba(255,255,255,0.12); border-radius: 8px;">
                </div>
                <div class="row">
                    <div class="col-sm-6 form-group">
                        <label for="author">Author</label>
                        <select name="author" id="author" class="form-control" style="background: #0f161e; color: #fff; border-color: #333;">
                            <option value="">-- Any Author --</option>
                            {% for a in authors %}
                            <option value="{{ a.ID }}">{{ a.Name }}</option>
                            {% endfor %}
                        </select>
                    </div>
                    <div class="col-sm-6 form-group">
                        <label for="series">Series</label>
                        <select name="series" id="series" class="form-control" style="background: #0f161e; color: #fff; border-color: #333;">
                            <option value="">-- Any Series --</option>
                            {% for s in series %}
                            <option value="{{ s.ID }}">{{ s.Name }}</option>
                            {% endfor %}
                        </select>
                    </div>
                </div>
                <div class="row">
                    <div class="col-sm-6 form-group">
                        <label for="tag">Category / Tag</label>
                        <select name="tag" id="tag" class="form-control" style="background: #0f161e; color: #fff; border-color: #333;">
                            <option value="">-- Any Category --</option>
                            {% for t in tags %}
                            <option value="{{ t.ID }}">{{ t.Name }}</option>
                            {% endfor %}
                        </select>
                    </div>
                    <div class="col-sm-6 form-group">
                        <label for="format">Format</label>
                        <select name="format" id="format" class="form-control" style="background: #0f161e; color: #fff; border-color: #333;">
                            <option value="">-- Any Format --</option>
                            {% for f in formats %}
                            <option value="{{ f.Name }}">{{ f.Name }}</option>
                            {% endfor %}
                        </select>
                    </div>
                </div>
                <div style="margin-top: 25px;">
                    <button type="submit" class="btn btn-primary btn-block btn-lg"><span class="glyphicon glyphicon-search"></span> Search Library</button>
                </div>
            </form>
        </div>
    </div>
    <script src="/static/js/libs/jquery.min.js"></script>
    <script src="/static/js/libs/bootstrap.min.js"></script>
</body>
</html>`

const adminHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Calibre-Web Automated | Admin</title>
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
    <div id="loader" hidden="true"></div>
    <div class="navbar navbar-default navbar-static-top" role="navigation">
        <div class="container-fluid">
            <div class="navbar-header">
                <a class="navbar-brand" href="/"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#38bdf8" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 8px; flex-shrink: 0;"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path></svg>Calibre-Web Automated <span class="label label-warning" style="font-size: 10px; vertical-align: middle; margin-left: 5px;">CWA</span></a>
            </div>
            <form class="navbar-form navbar-left" role="search" action="/search" method="GET">
                <div class="form-group input-group input-group-sm">
                    <input type="text" class="form-control" id="query_top" name="query" placeholder="Search Library">
                </div>
            </form>
            <div class="navbar-collapse collapse">
                <ul class="nav navbar-nav navbar-right" id="main-nav">
                    <li><a href="/"><span class="glyphicon glyphicon-book"></span> Library</a></li>
                    <li><a href="/tasks"><span class="glyphicon glyphicon-tasks"></span> Tasks</a></li>
                    <li class="active"><a href="/admin"><span class="glyphicon glyphicon-dashboard"></span> Admin</a></li>
                    <li><a href="/cwa-settings"><span class="glyphicon glyphicon-cog"></span> CWA Settings</a></li>
                    <li><a href="/cwa-stats"><span class="glyphicon glyphicon-stats"></span> Deep Analytics</a></li>
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
                    <span class="glyphicon glyphicon-dashboard" style="color: #38bdf8;"></span>
                    Calibre-Web Automated Admin
                </h1>
                <p class="cwa-page-subtitle">Server overview, automated operations, and user administration.</p>
            </div>
            <div>
                <span class="cwa-badge cwa-badge-info">CWA Engine</span>
            </div>
        </div>

        <div class="cwa-panel">
            <div class="cwa-metrics-grid">
                <div class="cwa-metric-box">
                    <div class="cwa-metric-label">Total Books</div>
                    <div class="cwa-metric-value cyan">{{ totalBooks }}</div>
                </div>
                <div class="cwa-metric-box">
                    <div class="cwa-metric-label">Authors</div>
                    <div class="cwa-metric-value green">{{ totalAuthors }}</div>
                </div>
                <div class="cwa-metric-box">
                    <div class="cwa-metric-label">Series</div>
                    <div class="cwa-metric-value amber">{{ totalSeries }}</div>
                </div>
            </div>

            <h2 class="cwa-panel-title">
                <span class="glyphicon glyphicon-cog"></span>
                Automated Operations
            </h2>
            <div style="margin-bottom: 30px;">
                <a href="/cwa-settings" class="cwa-op-card">
                    <div>
                        <div class="op-title"><span class="glyphicon glyphicon-cog" style="color: #38bdf8; margin-right: 8px;"></span> CWA Settings</div>
                        <div class="op-desc">Configure automated background services, target formats &amp; cover enforcement</div>
                    </div>
                    <span class="glyphicon glyphicon-chevron-right" style="color: #64748b; font-size: 12px;"></span>
                </a>
                <a href="/duplicates" class="cwa-op-card">
                    <div>
                        <div class="op-title"><span class="glyphicon glyphicon-duplicate" style="color: #fbbf24; margin-right: 8px;"></span> Smart Duplicate Detection &amp; Management</div>
                        <div class="op-desc">Scan, merge and deduplicate books with fuzzy title, author and ISBN matching</div>
                    </div>
                    <span class="glyphicon glyphicon-chevron-right" style="color: #64748b; font-size: 12px;"></span>
                </a>
                <a href="/cwa-convert-library-overview" class="cwa-op-card">
                    <div>
                        <div class="op-title"><span class="glyphicon glyphicon-refresh" style="color: #34d399; margin-right: 8px;"></span> Convert Library Tool</div>
                        <div class="op-desc">Multi-format conversion engine for EPUB, KEPUB, MOBI, AZW3 and PDF</div>
                    </div>
                    <span class="glyphicon glyphicon-chevron-right" style="color: #64748b; font-size: 12px;"></span>
                </a>
                <a href="/cwa-epub-fixer-overview" class="cwa-op-card">
                    <div>
                        <div class="op-title"><span class="glyphicon glyphicon-wrench" style="color: #f87171; margin-right: 8px;"></span> Automatic EPUB Fixer Service 2.0</div>
                        <div class="op-desc">Kindle Send-to-Kindle compliance, XML sanitization, and encoding repair</div>
                    </div>
                    <span class="glyphicon glyphicon-chevron-right" style="color: #64748b; font-size: 12px;"></span>
                </a>
                <a href="/cwa-stats" class="cwa-op-card">
                    <div>
                        <div class="op-title"><span class="glyphicon glyphicon-stats" style="color: #38bdf8; margin-right: 8px;"></span> Deep Stats &amp; Analytics</div>
                        <div class="op-desc">Real-time library metrics, CSV exports and server telemetry</div>
                    </div>
                    <span class="glyphicon glyphicon-chevron-right" style="color: #64748b; font-size: 12px;"></span>
                </a>
                <a href="/cwa-backup/list" class="cwa-op-card">
                    <div>
                        <div class="op-title"><span class="glyphicon glyphicon-floppy-save" style="color: #34d399; margin-right: 8px;"></span> Automated Back Up Service</div>
                        <div class="op-desc">Compressed Deflate archives of metadata.db and application state</div>
                    </div>
                    <span class="glyphicon glyphicon-chevron-right" style="color: #64748b; font-size: 12px;"></span>
                </a>
                <a href="/tasks" class="cwa-op-card">
                    <div>
                        <div class="op-title"><span class="glyphicon glyphicon-tasks" style="color: #94a3b8; margin-right: 8px;"></span> Background Tasks Monitor</div>
                        <div class="op-desc">View background workers and active ingestion watcher tasks</div>
                    </div>
                    <span class="glyphicon glyphicon-chevron-right" style="color: #64748b; font-size: 12px;"></span>
                </a>
            </div>

            <h2 class="cwa-panel-title">
                <span class="glyphicon glyphicon-user"></span>
                User Administration
            </h2>
            <table class="table cwa-table">
                <thead>
                    <tr>
                        <th>ID</th>
                        <th>Username</th>
                        <th>Email</th>
                        <th>Role</th>
                        <th>Admin Access</th>
                    </tr>
                </thead>
                <tbody>
                    {% for u in users %}
                    <tr>
                        <td>{{ u.ID }}</td>
                        <td><strong>{{ u.Name }}</strong></td>
                        <td>{{ u.Email }}</td>
                        <td>{{ u.Role }}</td>
                        <td>{% if u.IsAdmin %}<span class="cwa-badge cwa-badge-info">Admin</span>{% else %}<span class="cwa-badge cwa-badge-warning">User</span>{% endif %}</td>
                    </tr>
                    {% endfor %}
                </tbody>
            </table>
        </div>
    </div>
    <script src="/static/js/libs/jquery.min.js"></script>
    <script src="/static/js/libs/bootstrap.min.js"></script>
</body>
</html>`

const tasksHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Calibre-Web | Tasks</title>
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
    <div id="loader" hidden="true"></div>
    <div class="navbar navbar-default navbar-static-top" role="navigation">
        <div class="container-fluid">
            <div class="navbar-header">
                <a class="navbar-brand" href="/"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#38bdf8" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 8px; flex-shrink: 0;"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path></svg>Calibre-Web Automated <span class="label label-warning" style="font-size: 10px; vertical-align: middle; margin-left: 5px;">CWA</span></a>
            </div>
            <form class="navbar-form navbar-left" role="search" action="/search" method="GET">
                <div class="form-group input-group input-group-sm">
                    <input type="text" class="form-control" id="query_top" name="query" placeholder="Search Library">
                </div>
            </form>
            <div class="navbar-collapse collapse">
                <ul class="nav navbar-nav navbar-right" id="main-nav">
                    <li><a href="/"><span class="glyphicon glyphicon-book"></span> Library</a></li>
                    <li class="active"><a href="/tasks"><span class="glyphicon glyphicon-tasks"></span> Tasks</a></li>
                    <li><a href="/admin"><span class="glyphicon glyphicon-dashboard"></span> Admin</a></li>
                    <li><a href="/cwa-settings"><span class="glyphicon glyphicon-cog"></span> CWA Settings</a></li>
                    <li><a href="/cwa-stats"><span class="glyphicon glyphicon-stats"></span> Deep Analytics</a></li>
                </ul>
            </div>
        </div>
    </div>

    <div class="container" style="max-width: 1000px; margin-top: 76px; margin-bottom: 60px;">
        <div style="margin-bottom: 18px;">
            <a href="/" class="btn btn-default cwa-back-btn"><span class="glyphicon glyphicon-arrow-left"></span> Back to Library</a>
        </div>
        <div class="panel panel-default" style="background: rgba(18, 25, 38, 0.85); border-color: rgba(255,255,255,0.08); border-radius: 14px; padding: 32px; color: #fff; box-shadow: 0 16px 40px rgba(0,0,0,0.4); backdrop-filter: blur(16px);">
            <h2 style="margin-top: 0; margin-bottom: 20px; border-bottom: 1px solid rgba(255,255,255,0.08); padding-bottom: 14px; font-size: 22px; font-weight: 700; letter-spacing: -0.3px;">Background Tasks</h2>
            <div class="alert alert-info" style="background: rgba(56, 189, 248, 0.1); border-color: rgba(56, 189, 248, 0.25); color: #38bdf8; border-radius: 8px;">
                <span class="glyphicon glyphicon-info-sign"></span> Ingestion watcher is currently active with automated fsnotify event processing.
            </div>
            <table class="table table-hover" style="color: #ddd;">
                <thead>
                    <tr>
                        <th>Task</th>
                        <th>User</th>
                        <th>Status</th>
                        <th>Progress</th>
                        <th>Run Time</th>
                    </tr>
                </thead>
                <tbody>
                    <tr>
                        <td>Library Ingestion Watcher</td>
                        <td>admin</td>
                        <td><span class="label label-success">Active</span></td>
                        <td>100%</td>
                        <td>Continuous</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
    <script src="/static/js/libs/jquery.min.js"></script>
    <script src="/static/js/libs/bootstrap.min.js"></script>
</body>
</html>`

const statsHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Calibre-Web Automated | Deep Stats & Analytics</title>
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
    <div id="loader" hidden="true"></div>
    <div class="navbar navbar-default navbar-static-top" role="navigation">
        <div class="container-fluid">
            <div class="navbar-header">
                <a class="navbar-brand" href="/"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#38bdf8" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 8px; flex-shrink: 0;"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path></svg>Calibre-Web Automated <span class="label label-warning" style="font-size: 10px; vertical-align: middle; margin-left: 5px;">CWA</span></a>
            </div>
            <form class="navbar-form navbar-left" role="search" action="/search" method="GET">
                <div class="form-group input-group input-group-sm">
                    <input type="text" class="form-control" id="query_top" name="query" placeholder="Search Library">
                </div>
            </form>
            <div class="navbar-collapse collapse">
                <ul class="nav navbar-nav navbar-right" id="main-nav">
                    <li><a href="/"><span class="glyphicon glyphicon-book"></span> Library</a></li>
                    <li><a href="/tasks"><span class="glyphicon glyphicon-tasks"></span> Tasks</a></li>
                    <li><a href="/admin"><span class="glyphicon glyphicon-dashboard"></span> Admin</a></li>
                    <li><a href="/cwa-settings"><span class="glyphicon glyphicon-cog"></span> CWA Settings</a></li>
                    <li class="active"><a href="/cwa-stats"><span class="glyphicon glyphicon-stats"></span> Deep Analytics</a></li>
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
                    <span class="glyphicon glyphicon-stats" style="color: #38bdf8;"></span>
                    Deep Stats &amp; Analytics
                </h1>
                <p class="cwa-page-subtitle">Real-time library metrics, CSV export center, and runtime telemetry.</p>
            </div>
            <div>
                <span class="cwa-badge cwa-badge-success">Live Telemetry</span>
            </div>
        </div>

        <div class="cwa-panel">
            <div class="cwa-metrics-grid">
                <div class="cwa-metric-box">
                    <div class="cwa-metric-label">Total Books</div>
                    <div class="cwa-metric-value cyan">{{ totalBooks }}</div>
                </div>
                <div class="cwa-metric-box">
                    <div class="cwa-metric-label">Authors</div>
                    <div class="cwa-metric-value green">{{ totalAuthors }}</div>
                </div>
                <div class="cwa-metric-box">
                    <div class="cwa-metric-label">Series</div>
                    <div class="cwa-metric-value amber">{{ totalSeries }}</div>
                </div>
            </div>

            <h2 class="cwa-panel-title">
                <span class="glyphicon glyphicon-download-alt"></span>
                CSV Analytics Export Center
            </h2>
            <p class="cwa-panel-description">Download comprehensive tabular reports of your collection for external analytics or spreadsheet auditing:</p>
            <div class="cwa-actions-bar" style="margin-bottom: 30px;">
                <a href="/cwa-stats-export-csv/library" class="btn btn-default"><span class="glyphicon glyphicon-file"></span> Library Overview CSV</a>
                <a href="/cwa-stats-export-csv/books" class="btn btn-default"><span class="glyphicon glyphicon-book"></span> Books Details CSV</a>
                <a href="/cwa-stats-export-csv/authors" class="btn btn-default"><span class="glyphicon glyphicon-user"></span> Authors Summary CSV</a>
                <a href="/cwa-stats-export-csv/series" class="btn btn-default"><span class="glyphicon glyphicon-list"></span> Series Summary CSV</a>
            </div>

            <h2 class="cwa-panel-title">
                <span class="glyphicon glyphicon-dashboard"></span>
                Server Health &amp; Resource Footprint
            </h2>
            <div class="cwa-metrics-grid" id="server-health-stats">
                <div class="cwa-metric-box">
                    <div class="cwa-metric-label">Heap Allocated</div>
                    <div id="heap-alloc" class="cwa-metric-value cyan">Loading...</div>
                </div>
                <div class="cwa-metric-box">
                    <div class="cwa-metric-label">System RAM</div>
                    <div id="heap-sys" class="cwa-metric-value green">Loading...</div>
                </div>
                <div class="cwa-metric-box">
                    <div class="cwa-metric-label">Goroutines</div>
                    <div id="goroutines" class="cwa-metric-value amber">Loading...</div>
                </div>
                <div class="cwa-metric-box">
                    <div class="cwa-metric-label">RAM Target (Limit)</div>
                    <div class="cwa-metric-value green">&lt; 40 MB</div>
                </div>
            </div>

            <h2 class="cwa-panel-title" style="margin-top: 30px;">
                <span class="glyphicon glyphicon-folder-open"></span>
                Available Formats in Library
            </h2>
            <div class="row">
                {% for f in formats %}
                <div class="col-sm-4 col-xs-6" style="margin-bottom: 10px;">
                    <div style="background: #0d151e; border: 1px solid rgba(255,255,255,0.06); border-radius: 6px; padding: 12px 16px; display: flex; justify-content: space-between; align-items: center;">
                        <span style="font-weight: 600; color: #f1f5f9;">{{ f.Name }}</span>
                        <span class="cwa-badge cwa-badge-info">{{ f.Count }}</span>
                    </div>
                </div>
                {% endfor %}
            </div>
        </div>
    </div>
    <script src="/static/js/libs/jquery.min.js"></script>
    <script src="/static/js/libs/bootstrap.min.js"></script>
    <script>
    fetch('/health').then(r => r.json()).then(d => {
        document.getElementById('heap-alloc').innerText = d.heap_alloc || 'N/A';
        document.getElementById('heap-sys').innerText = d.heap_sys || 'N/A';
        document.getElementById('goroutines').innerText = d.goroutines || 'N/A';
    });
    </script>
</body>
</html>`

const duplicatesHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Calibre-Web Automated | Duplicates</title>
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
    <div id="loader" hidden="true"></div>
    <div class="navbar navbar-default navbar-static-top" role="navigation">
        <div class="container-fluid">
            <div class="navbar-header">
                <a class="navbar-brand" href="/"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#38bdf8" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 8px; flex-shrink: 0;"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path></svg>Calibre-Web Automated <span class="label label-warning" style="font-size: 10px; vertical-align: middle; margin-left: 5px;">CWA</span></a>
            </div>
            <form class="navbar-form navbar-left" role="search" action="/search" method="GET">
                <div class="form-group input-group input-group-sm">
                    <input type="text" class="form-control" id="query_top" name="query" placeholder="Search Library">
                </div>
            </form>
            <div class="navbar-collapse collapse">
                <ul class="nav navbar-nav navbar-right" id="main-nav">
                    <li><a href="/"><span class="glyphicon glyphicon-book"></span> Library</a></li>
                    <li><a href="/tasks"><span class="glyphicon glyphicon-tasks"></span> Tasks</a></li>
                    <li><a href="/admin"><span class="glyphicon glyphicon-dashboard"></span> Admin</a></li>
                    <li><a href="/cwa-settings"><span class="glyphicon glyphicon-cog"></span> CWA Settings</a></li>
                    <li class="active"><a href="/duplicates"><span class="glyphicon glyphicon-duplicate"></span> Duplicates</a></li>
                </ul>
            </div>
        </div>
    </div>
    <div class="cwa-container" style="max-width: 1140px;">
        <div class="cwa-page-header">
            <div>
                <a href="/" class="btn btn-default cwa-back-btn" style="margin-bottom: 12px;">
                    <span class="glyphicon glyphicon-arrow-left"></span> Back to Library
                </a>
                <h1 class="cwa-page-title">
                    <span class="glyphicon glyphicon-duplicate" style="color: #38bdf8;"></span>
                    Smart Duplicate Detection &amp; Management
                </h1>
                <p class="cwa-page-subtitle">Identifies books with matching titles, authors, and external identifiers (ISBN/Goodreads).</p>
            </div>
            <div>
                <button type="button" class="btn btn-default" id="trigger_scan" onclick="triggerScan()">
                    <span class="glyphicon glyphicon-refresh"></span> Scan for Duplicates Now
                </button>
            </div>
        </div>

        {% if groups %}
        <div class="cwa-panel">
            <h2 class="cwa-panel-title">
                <span class="glyphicon glyphicon-cog"></span>
                Automated Resolution
            </h2>
            <div class="cwa-actions-bar">
                <label for="strategy" class="cwa-label" style="margin: 0;">Strategy:</label>
                <select id="strategy" class="form-control cwa-select" style="width: auto; min-width: 280px;">
                    <option value="newest">Keep Newest (Retain most recently added, delete older)</option>
                    <option value="oldest">Keep Oldest (Retain earliest added, delete newer)</option>
                    <option value="merge">Merge Formats (Combine all unique formats into primary)</option>
                    <option value="most_metadata">Keep Most Metadata (Retain copy with richest synopsis &amp; tags)</option>
                </select>
                <button type="button" class="btn btn-default" onclick="previewResolution()">
                    <span class="glyphicon glyphicon-eye-open"></span> Preview
                </button>
                <button type="button" class="btn btn-danger" onclick="executeResolution()">
                    <span class="glyphicon glyphicon-flash"></span> Execute Resolution
                </button>
            </div>
        </div>

        {% for g in groups %}
        <div class="cwa-panel" id="group-{{ g.Hash }}" style="margin-bottom: 25px;">
            <div style="display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid rgba(255,255,255,0.08); padding-bottom: 12px; margin-bottom: 18px;">
                <div>
                    <h3 style="margin: 0; font-size: 18px; font-weight: 600; color: #f1f5f9;">{{ g.Title }}</h3>
                    <div style="margin-top: 5px;">
                        <span class="cwa-badge cwa-badge-warning">{{ g.Confidence }}% Match Confidence</span>
                        <span class="cwa-badge cwa-badge-info" style="margin-left: 6px;">{{ g.Books|length }} Copies Found</span>
                    </div>
                </div>
                <div>
                    <button type="button" class="btn btn-sm btn-default" onclick="dismissGroup('{{ g.Hash }}')">
                        <span class="glyphicon glyphicon-eye-close"></span> Dismiss Group
                    </button>
                </div>
            </div>

            <div class="row">
                {% for b in g.Books %}
                <div class="col-sm-6" style="margin-bottom: 15px;">
                    <div style="background: #0d151e; border: 1px solid rgba(255,255,255,0.06); border-radius: 8px; padding: 15px; display: flex; gap: 15px;">
                        <div style="flex-shrink: 0;">
                            <a href="/book/{{ b.ID }}">
                                <img src="/cover/{{ b.ID }}" alt="{{ b.Title }}" style="width: 75px; height: 110px; object-fit: cover; border-radius: 4px; box-shadow: 0 2px 8px rgba(0,0,0,0.4);" loading="lazy">
                            </a>
                        </div>
                        <div style="flex: 1; min-width: 0; font-size: 12px; color: #cbd5e1;">
                            <div style="font-weight: 600; font-size: 14px; margin-bottom: 4px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">
                                <a href="/book/{{ b.ID }}" style="color: #f1f5f9;">{{ b.Title }}</a>
                            </div>
                            <div style="color: #8194a7; margin-bottom: 3px;">
                                <strong>ID:</strong> #{{ b.ID }}
                            </div>
                            <div style="color: #8194a7; margin-bottom: 3px;">
                                <strong>Authors:</strong> {% for a in b.Authors %}{{ a }}{% if not forloop.Last %}, {% endif %}{% endfor %}
                            </div>
                            <div style="color: #8194a7; margin-bottom: 10px;">
                                <strong>Formats:</strong> {% for f in b.Formats %}{{ f.Format }} ({{ f.Size|filesizeformat }}){% if not forloop.Last %}, {% endif %}{% endfor %}
                            </div>
                            <div class="cwa-actions-bar">
                                <a href="/admin/book/{{ b.ID }}" class="btn btn-xs btn-primary">
                                    <span class="glyphicon glyphicon-pencil"></span> Edit
                                </a>
                                <button type="button" class="btn btn-xs btn-danger" onclick="if(confirm('Delete book #{{ b.ID }}?')){fetch('/delete/{{ b.ID }}',{method:'POST'}).then(function(){location.reload();});}">
                                    <span class="glyphicon glyphicon-trash"></span> Delete
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
                {% endfor %}
            </div>
        </div>
        {% endfor %}
        {% else %}
        <div class="cwa-panel text-center" style="padding: 50px 20px;">
            <span class="glyphicon glyphicon-ok-sign" style="font-size: 42px; color: #4ade80; margin-bottom: 15px; display: block;"></span>
            <h3 style="margin-top: 0; color: #f1f5f9; font-weight: 600;">No Duplicates Detected</h3>
            <p style="color: #8194a7; max-width: 500px; margin: 0 auto;">Your library is completely deduplicated. No matching titles, authors, or external identifiers were found.</p>
        </div>
        {% endif %}
    </div>

    <script src="/static/js/libs/jquery.min.js"></script>
    <script src="/static/js/libs/bootstrap.min.js"></script>
    <script>
    function triggerScan() {
        fetch('/duplicates/trigger-scan', {method: 'POST'})
            .then(r => r.json())
            .then(data => { alert('Scan complete.'); location.reload(); })
            .catch(err => alert('Scan error: ' + err));
    }
    function dismissGroup(hash) {
        fetch('/duplicates/dismiss/' + hash, {method: 'POST'})
            .then(r => r.json())
            .then(data => { $('#group-' + hash).fadeOut(); });
    }
    function previewResolution() {
        const strat = document.getElementById('strategy').value;
        fetch('/duplicates/preview-resolution', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({strategy: strat})
        }).then(r => r.json()).then(data => {
            alert('Resolution preview: Strategy "' + strat + '" is ready to execute.');
        });
    }
    function executeResolution() {
        const strat = document.getElementById('strategy').value;
        if (!confirm('Permanently resolve duplicates using strategy "' + strat + '"?')) return;
        fetch('/duplicates/execute-resolution', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({strategy: strat})
        }).then(r => r.json()).then(data => {
            alert('Duplicates resolved.');
            location.reload();
        });
    }
    </script>
</body>
</html>`

const bookEditHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Calibre-Web | Edit {{ book.Title }}</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link href="/static/css/libs/bootstrap.min.css" rel="stylesheet" media="screen">
    <link href="/static/css/style.css" rel="stylesheet" media="screen">
    <link href="/static/css/caliBlur.css" rel="stylesheet" media="screen">
    <link href="/static/css/caliBlur_override.css" rel="stylesheet" media="screen">
    <link href="/static/css/layout-refresh.css" rel="stylesheet" media="screen">
    <link href="/static/css/cwa.css?v=2026.5" rel="stylesheet" media="screen">
    <style>
    .edit-container { max-width: 1200px; margin: 30px auto; padding: 0 15px; }
    .edit-card { background: #18222d; border-radius: 12px; border: 1px solid rgba(255,255,255,0.08); padding: 30px; color: #fff; }
    .cover-section { text-align: center; margin-bottom: 25px; }
    .cover-section img { max-width: 100%; height: auto; border-radius: 8px; box-shadow: 0 8px 16px rgba(0,0,0,0.5); }
    .form-control { background: #0f161e !important; color: #fff !important; border-color: #333 !important; }
    .form-control:focus { border-color: #5bc0de !important; }
    label { color: #ddd; font-weight: 600; margin-top: 10px; }
    .btn-file { position: relative; overflow: hidden; }
    .btn-file input[type=file] { position: absolute; top: 0; right: 0; min-width: 100%; min-height: 100%; font-size: 100px; text-align: right; filter: alpha(opacity=0); opacity: 0; outline: none; background: white; cursor: inherit; display: block; }
    </style>
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
    <div id="loader" hidden="true"></div>
    <div class="edit-container">
        <div style="margin-bottom: 20px;">
            <a href="/book/{{ book.ID }}" class="btn btn-default cwa-back-btn"><span class="glyphicon glyphicon-arrow-left"></span> Back to Book</a>
        </div>
        <div class="edit-card">
            <div class="row">
                <!-- Coluna da Esquerda: Capa, Formatos e Ações de Exclusão -->
                <div class="col-sm-4 col-md-3">
                    <div class="cover-section">
                        <img id="detailcover" src="/cover/{{ book.ID }}" alt="{{ book.Title }}">
                    </div>

                    <div style="margin-bottom: 25px;">
                        <button type="button" class="btn btn-danger btn-block" onclick="if(confirm('Delete entire book?')){fetch('/delete/{{ book.ID }}',{method:'POST'}).then(function(){window.location.href='/';});}"><span class="glyphicon glyphicon-trash"></span> Delete Book</button>
                    </div>

                    {% if book.Formats %}
                    <div class="panel panel-default" style="background: #0f161e; border-color: #333; color: #ddd;">
                        <div class="panel-heading" style="background: #1b242f; color: #fff; border-color: #333; font-weight: bold;">Formats</div>
                        <ul class="list-group" style="margin: 0;">
                            {% for f in book.Formats %}
                            <li class="list-group-item" style="background: transparent; border-color: #2b3e50; display: flex; justify-content: space-between; align-items: center;">
                                <span><strong>{{ f.Format }}</strong> ({{ f.Size|filesizeformat }})</span>
                                <div class="btn-group btn-group-xs">
                                    <a href="/download/{{ book.ID }}/{{ f.Format }}" class="btn btn-default"><span class="glyphicon glyphicon-download"></span></a>
                                    <button type="button" class="btn btn-danger" onclick="if(confirm('Delete {{ f.Format }} format?')){fetch('/delete/{{ book.ID }}/{{ f.Format }}',{method:'POST'}).then(function(){location.reload();});}"><span class="glyphicon glyphicon-remove"></span></button>
                                </div>
                            </li>
                            {% endfor %}
                        </ul>
                    </div>
                    {% endif %}

                    <!-- Upload de Novo Formato -->
                    <div style="margin-top: 15px;">
                        <form action="/upload" method="POST" enctype="multipart/form-data">
                            <span class="btn btn-primary btn-block btn-file">
                                <span class="glyphicon glyphicon-upload"></span> Add Format File
                                <input type="file" name="file" onchange="this.form.submit()">
                            </span>
                        </form>
                    </div>
                </div>

                <!-- Coluna da Direita: Formulário Completo de Metadados -->
                <div class="col-sm-8 col-md-9">
                    <h2 style="margin-top: 0; margin-bottom: 20px; border-bottom: 1px solid rgba(255,255,255,0.1); padding-bottom: 10px;">Edit Book Metadata</h2>
                    <form role="form" action="/admin/book/{{ book.ID }}" method="POST" enctype="multipart/form-data">
                        <div class="row">
                            <div class="col-sm-8 form-group">
                                <label for="title">Book Title</label>
                                <input type="text" class="form-control" name="title" id="title" value="{{ book.Title }}" required>
                            </div>
                            <div class="col-sm-4 form-group">
                                <label for="sort">Title Sort</label>
                                <input type="text" class="form-control" name="sort" id="sort" value="{{ book.Sort }}">
                            </div>
                        </div>

                        <div class="form-group">
                            <label for="authors">Authors (comma separated)</label>
                            <input type="text" class="form-control" name="authors" id="authors" value="{% for a in book.Authors %}{{ a }}{% if not forloop.Last %}, {% endif %}{% endfor %}">
                        </div>

                        <div class="row">
                            <div class="col-sm-8 form-group">
                                <label for="series">Series</label>
                                <input type="text" class="form-control" name="series" id="series" value="{{ book.Series }}">
                            </div>
                            <div class="col-sm-4 form-group">
                                <label for="series_index">Series Index</label>
                                <input type="number" step="0.01" min="0" class="form-control" name="series_index" id="series_index" value="{{ book.SeriesIndex }}">
                            </div>
                        </div>

                        <div class="row">
                            <div class="col-sm-6 form-group">
                                <label for="tags">Tags / Categories (comma separated)</label>
                                <input type="text" class="form-control" name="tags" id="tags" value="{% for t in book.Tags %}{{ t }}{% if not forloop.Last %}, {% endif %}{% endfor %}">
                            </div>
                            <div class="col-sm-3 form-group">
                                <label for="rating">Rating (0 - 10)</label>
                                <input type="number" min="0" max="10" step="1" class="form-control" name="rating" id="rating" value="{{ book.Rating }}">
                            </div>
                            <div class="col-sm-3 form-group">
                                <label for="languages">Language</label>
                                <input type="text" class="form-control" name="languages" id="languages" value="{% for l in book.Languages %}{{ l }}{% if not forloop.Last %}, {% endif %}{% endfor %}">
                            </div>
                        </div>

                        <div class="form-group">
                            <label for="publisher">Publisher</label>
                            <input type="text" class="form-control" name="publisher" id="publisher" value="{% for p in book.Publishers %}{{ p }}{% if not forloop.Last %}, {% endif %}{% endfor %}">
                        </div>

                        <div class="form-group">
                            <label for="comments">Synopsis / Description</label>
                            <textarea class="form-control" name="comments" id="comments" rows="6">{{ book.Comment }}</textarea>
                        </div>

                        <div class="form-group">
                            <label for="cover_file">Upload New Cover Image</label>
                            <input type="file" class="form-control" name="cover_file" id="cover_file" accept=".jpg,.jpeg,.png,.webp">
                        </div>

                        <div style="margin-top: 30px; display: flex; gap: 12px;">
                            <button type="submit" class="btn btn-success btn-lg"><span class="glyphicon glyphicon-ok"></span> Save Changes</button>
                            <a href="/book/{{ book.ID }}" class="btn btn-default btn-lg">Cancel</a>
                        </div>
                    </form>
                </div>
            </div>
        </div>
    </div>
    <script src="/static/js/libs/jquery.min.js"></script>
    <script src="/static/js/libs/bootstrap.min.js"></script>
</body>
</html>`

const shelfEditHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Calibre-Web | {{ title }}</title>
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
    <div id="loader" hidden="true"></div>
    <div class="container" style="max-width: 600px; padding-top: 40px;">
        <div style="margin-bottom: 20px;">
            <a href="/" class="btn btn-default cwa-back-btn"><span class="glyphicon glyphicon-arrow-left"></span> Back</a>
        </div>
        <div class="panel panel-default" style="background: #18222d; border-color: rgba(255,255,255,0.1); border-radius: 12px; padding: 30px; color: #fff;">
            <h2 style="margin-top: 0; margin-bottom: 20px; border-bottom: 1px solid rgba(255,255,255,0.1); padding-bottom: 10px;">{{ title }}</h2>
            <form role="form" method="POST">
                <div class="form-group">
                    <label for="title" style="color: #ddd;">Shelf Name</label>
                    <input type="text" class="form-control" name="title" id="title" value="{{ shelf.Name }}" required style="background: #0f161e; color: #fff; border-color: #333;">
                </div>
                <div class="checkbox" style="margin-bottom: 25px;">
                    <label style="color: #bbb;">
                        <input type="checkbox" name="is_public" {% if shelf.IsPublic %}checked{% endif %}> Share with Everyone (Public Shelf)
                    </label>
                </div>
                <button type="submit" class="btn btn-primary"><span class="glyphicon glyphicon-ok"></span> Save Shelf</button>
                {% if shelf.ID > 0 %}
                <a href="/shelf/{{ shelf.ID }}" class="btn btn-default">Cancel</a>
                {% else %}
                <a href="/" class="btn btn-default">Cancel</a>
                {% endif %}
            </form>
        </div>
    </div>
    <script src="/static/js/libs/jquery.min.js"></script>
    <script src="/static/js/libs/bootstrap.min.js"></script>
</body>
</html>`

const loginHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Calibre-Web Automated | Sign In</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link href="/static/css/libs/bootstrap.min.css" rel="stylesheet" media="screen">
    <link href="/static/css/style.css" rel="stylesheet" media="screen">
    <link href="/static/css/caliBlur.css" rel="stylesheet" media="screen">
    <link href="/static/css/cwa.css?v=2026.5" rel="stylesheet" media="screen">
</head>
<body class="blur" style="display: flex; justify-content: center; align-items: center; min-height: 100vh; background: radial-gradient(circle at 50% 30%, rgba(30, 41, 59, 0.6) 0%, #0b0f19 80%) !important;">
    <div class="cwa-login" style="width: 100%; max-width: 400px; padding: 20px;">
        <div class="cwa-login-card" style="background: rgba(18, 25, 38, 0.85); padding: 40px 36px; border-radius: 16px; border: 1px solid rgba(255,255,255,0.08); box-shadow: 0 20px 45px rgba(0,0,0,0.5); backdrop-filter: blur(20px); -webkit-backdrop-filter: blur(20px); color: #fff;">
            <div class="cwa-login-brand text-center" style="margin-bottom: 28px;">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="#38bdf8" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-bottom: 12px;"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path></svg>
                <h2 style="margin: 0; font-size: 20px; font-weight: 700; letter-spacing: -0.3px;">Calibre-Web Automated</h2>
                <p style="color: #64748b; font-size: 13px; margin-top: 4px; margin-bottom: 0;">Sign in to your library</p>
            </div>

            <form method="POST" action="/login" class="cwa-login-form">
                <div class="form-group" style="margin-bottom: 18px;">
                    <label for="username" style="font-weight: 600; font-size: 13px; color: #cbd5e1; margin-bottom: 6px;">Username</label>
                    <input type="text" id="username" name="username" class="form-control" value="admin" required autofocus style="background: rgba(255,255,255,0.05); border: 1px solid rgba(255,255,255,0.12); color: #fff; height: 42px; border-radius: 8px; padding: 8px 14px; font-size: 13px;">
                </div>

                <div class="form-group" style="margin-bottom: 22px;">
                    <label for="password" style="font-weight: 600; font-size: 13px; color: #cbd5e1; margin-bottom: 6px;">Password</label>
                    <input type="password" id="password" name="password" class="form-control" value="admin123" required style="background: rgba(255,255,255,0.05); border: 1px solid rgba(255,255,255,0.12); color: #fff; height: 42px; border-radius: 8px; padding: 8px 14px; font-size: 13px;">
                </div>

                <div class="form-group" style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px;">
                    <label style="font-weight: normal; color: #94a3b8; font-size: 13px; margin: 0; cursor: pointer;">
                        <input type="checkbox" name="remember_me" checked style="accent-color: #3b82f6; margin-right: 6px; vertical-align: middle;"> Remember Me
                    </label>
                </div>

                <button type="submit" class="btn btn-primary btn-block" style="height: 44px; border-radius: 8px; font-weight: 600; font-size: 14px;">Sign In</button>
            </form>
        </div>
    </div>
</body>
</html>`

const readerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <title>{{ title }} — CWA NextGen Reader</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <link rel="apple-touch-icon" sizes="140x140" href="/static/favicon.ico">
    <link rel="shortcut icon" href="/static/favicon.ico">
    <script src="/static/js/libs/jquery.min.js"></script>
    <script src="/static/js/compress/jszip_epub.min.js"></script>
    <script src="/static/js/libs/epub.min.js"></script>
    <style>
        :root {
            --bg-color: #121214;
            --text-color: #e4e4e7;
            --hud-bg: rgba(24, 24, 27, 0.88);
            --border-color: rgba(255, 255, 255, 0.08);
            --accent-color: #38bdf8;
            --sidebar-bg: #18181b;
            --btn-hover: rgba(255, 255, 255, 0.08);
            --font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            --font-size: 100%;
        }

        body[data-theme="dark"] {
            --bg-color: #121214;
            --text-color: #e4e4e7;
            --hud-bg: rgba(24, 24, 27, 0.92);
            --border-color: rgba(255, 255, 255, 0.1);
            --accent-color: #38bdf8;
            --sidebar-bg: #18181b;
            --btn-hover: rgba(255, 255, 255, 0.1);
        }

        body[data-theme="oled"] {
            --bg-color: #000000;
            --text-color: #d1d5db;
            --hud-bg: rgba(0, 0, 0, 0.95);
            --border-color: rgba(255, 255, 255, 0.15);
            --accent-color: #38bdf8;
            --sidebar-bg: #09090b;
            --btn-hover: rgba(255, 255, 255, 0.12);
        }

        body[data-theme="sepia"] {
            --bg-color: #f4ecd8;
            --text-color: #423528;
            --hud-bg: rgba(238, 228, 207, 0.95);
            --border-color: rgba(66, 53, 40, 0.12);
            --accent-color: #8b5cf6;
            --sidebar-bg: #eae0c8;
            --btn-hover: rgba(66, 53, 40, 0.08);
        }

        body[data-theme="light"] {
            --bg-color: #ffffff;
            --text-color: #18181b;
            --hud-bg: rgba(255, 255, 255, 0.94);
            --border-color: rgba(0, 0, 0, 0.08);
            --accent-color: #0284c7;
            --sidebar-bg: #f8fafc;
            --btn-hover: rgba(0, 0, 0, 0.05);
        }

        * { box-sizing: border-box; }
        body, html {
            margin: 0; padding: 0; width: 100%; height: 100%;
            overflow: hidden; background: var(--bg-color); color: var(--text-color);
            font-family: var(--font-family); user-select: none;
            transition: background-color 0.25s ease, color 0.25s ease;
        }

        /* Top HUD */
        #top-hud {
            position: fixed; top: 0; left: 0; right: 0; height: 54px;
            background: var(--hud-bg); backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px);
            border-bottom: 1px solid var(--border-color); display: flex; align-items: center;
            justify-content: space-between; padding: 0 16px; z-index: 50;
            transition: transform 0.25s ease, opacity 0.25s ease;
        }
        #top-hud.hidden { transform: translateY(-100%); opacity: 0; pointer-events: none; }

        .hud-left, .hud-right { display: flex; align-items: center; gap: 8px; }
        .hud-center {
            position: absolute; left: 50%; transform: translateX(-50%);
            display: flex; flex-direction: column; align-items: center;
            max-width: 45%; pointer-events: none;
        }
        .hud-book-title { font-size: 13px; font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 100%; }
        .hud-chapter-title { font-size: 11px; opacity: 0.7; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 100%; }

        .btn-icon {
            background: transparent; border: 1px solid transparent; color: var(--text-color);
            border-radius: 8px; width: 36px; height: 36px; display: inline-flex; align-items: center;
            justify-content: center; cursor: pointer; transition: all 0.15s ease; text-decoration: none;
        }
        .btn-icon:hover { background: var(--btn-hover); border-color: var(--border-color); }
        .btn-icon svg { width: 18px; height: 18px; fill: none; stroke: currentColor; stroke-width: 2; stroke-linecap: round; stroke-linejoin: round; }

        .sync-badge {
            font-size: 11px; padding: 3px 8px; border-radius: 12px; background: rgba(56, 189, 248, 0.15);
            color: var(--accent-color); font-weight: 500; display: inline-flex; align-items: center; gap: 4px;
            transition: opacity 0.3s;
        }
        .sync-badge.saved { background: rgba(34, 197, 94, 0.15); color: #22c55e; }

        /* Viewer area */
        #viewer-container {
            position: absolute; top: 0; bottom: 0; left: 0; right: 0;
            display: flex; align-items: center; justify-content: center;
        }
        #viewer {
            width: 100%; height: 100%;
            max-width: 1400px; margin: 0 auto;
        }

        /* Click/Tap Navigation Zones */
        .tap-zone {
            position: absolute; top: 54px; bottom: 44px; width: 18%; z-index: 20;
            cursor: pointer; opacity: 0; transition: opacity 0.2s;
            display: flex; align-items: center; justify-content: center;
        }
        .tap-zone:hover { opacity: 1; }
        .tap-prev { left: 0; }
        .tap-next { right: 0; }
        .tap-btn {
            background: var(--hud-bg); color: var(--text-color); border: 1px solid var(--border-color);
            border-radius: 50%; width: 44px; height: 44px; display: flex; align-items: center; justify-content: center;
            box-shadow: 0 4px 12px rgba(0,0,0,0.3);
        }

        /* Bottom Floating HUD */
        #bottom-hud {
            position: fixed; bottom: 0; left: 0; right: 0; height: 44px;
            background: var(--hud-bg); backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px);
            border-top: 1px solid var(--border-color); display: flex; align-items: center;
            justify-content: space-between; padding: 0 20px; z-index: 50; font-size: 12px;
            transition: transform 0.25s ease, opacity 0.25s ease;
        }
        #bottom-hud.hidden { transform: translateY(100%); opacity: 0; pointer-events: none; }

        .progress-bar-wrap {
            flex: 1; max-width: 500px; height: 4px; background: rgba(255,255,255,0.1);
            border-radius: 2px; margin: 0 16px; overflow: hidden; position: relative; cursor: pointer;
        }
        .progress-bar-fill {
            height: 100%; background: var(--accent-color); width: 0%;
            transition: width 0.3s ease;
        }

        /* Left Drawer (Table of Contents) */
        #toc-drawer {
            position: fixed; top: 0; bottom: 0; left: 0; width: 340px; max-width: 85vw;
            background: var(--sidebar-bg); border-right: 1px solid var(--border-color);
            z-index: 100; transform: translateX(-100%); transition: transform 0.25s cubic-bezier(0.16, 1, 0.3, 1);
            display: flex; flex-direction: column; box-shadow: 4px 0 24px rgba(0,0,0,0.4);
        }
        #toc-drawer.open { transform: translateX(0); }
        .drawer-header {
            padding: 16px; border-bottom: 1px solid var(--border-color);
            display: flex; align-items: center; justify-content: space-between;
        }
        .drawer-header h3 { margin: 0; font-size: 15px; font-weight: 600; }
        .toc-search {
            padding: 10px 16px; border-bottom: 1px solid var(--border-color);
        }
        .toc-search input {
            width: 100%; background: rgba(255, 255, 255, 0.06); border: 1px solid var(--border-color);
            border-radius: 6px; padding: 8px 12px; color: var(--text-color); font-size: 13px; outline: none;
        }
        .toc-list {
            flex: 1; overflow-y: auto; padding: 8px 0; margin: 0; list-style: none;
        }
        .toc-item {
            padding: 10px 18px; font-size: 13px; cursor: pointer; border-left: 3px solid transparent;
            transition: background 0.15s, border-color 0.15s; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
        }
        .toc-item:hover { background: var(--btn-hover); }
        .toc-item.active { border-color: var(--accent-color); color: var(--accent-color); font-weight: 600; background: rgba(56, 189, 248, 0.08); }

        /* Settings Popover */
        #settings-popover {
            position: fixed; top: 60px; right: 16px; width: 320px;
            background: var(--sidebar-bg); border: 1px solid var(--border-color);
            border-radius: 12px; padding: 18px; z-index: 100; display: none;
            box-shadow: 0 12px 32px rgba(0,0,0,0.4); animation: fadeIn 0.15s ease;
        }
        #settings-popover.open { display: block; }
        @keyframes fadeIn { from { opacity: 0; transform: translateY(-8px); } to { opacity: 1; transform: translateY(0); } }

        .settings-section { margin-bottom: 16px; }
        .settings-label { font-size: 11px; text-transform: uppercase; letter-spacing: 0.05em; opacity: 0.7; margin-bottom: 8px; font-weight: 600; }
        .theme-buttons { display: grid; grid-template-columns: repeat(4, 1fr); gap: 6px; }
        .theme-btn {
            padding: 8px 4px; border-radius: 6px; border: 1px solid var(--border-color);
            cursor: pointer; font-size: 11px; font-weight: 500; text-align: center; background: rgba(255,255,255,0.04);
            color: var(--text-color); transition: all 0.15s;
        }
        .theme-btn.active { border-color: var(--accent-color); background: rgba(56, 189, 248, 0.15); color: var(--accent-color); }
        .font-btn-row { display: flex; gap: 8px; margin-top: 6px; }
        .font-size-ctrl {
            display: flex; align-items: center; justify-content: space-between;
            background: rgba(255,255,255,0.05); border: 1px solid var(--border-color);
            border-radius: 8px; padding: 4px 12px;
        }
        .ctrl-btn {
            background: transparent; border: none; color: var(--text-color); font-size: 16px;
            font-weight: bold; cursor: pointer; padding: 4px 10px; border-radius: 4px;
        }
        .ctrl-btn:hover { background: var(--btn-hover); }

        /* Backdrop overlay */
        #backdrop {
            position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 90;
            display: none; backdrop-filter: blur(2px);
        }
        #backdrop.open { display: block; }

        /* Loading Spinner */
        #loading {
            position: fixed; inset: 0; display: flex; flex-direction: column;
            align-items: center; justify-content: center; background: var(--bg-color); z-index: 200;
        }
        .spinner {
            width: 40px; height: 40px; border: 3px solid rgba(56, 189, 248, 0.2);
            border-top-color: var(--accent-color); border-radius: 50%;
            animation: spin 0.8s linear infinite;
        }
        @keyframes spin { to { transform: rotate(360deg); } }
    </style>
</head>
<body data-theme="dark">
    <!-- Loading Screen -->
    <div id="loading">
        <div class="spinner"></div>
        <p style="margin-top: 16px; font-size: 14px; opacity: 0.8;">A carregar livro...</p>
    </div>

    <!-- Backdrop -->
    <div id="backdrop"></div>

    <!-- Top Toolbar HUD -->
    <header id="top-hud">
        <div class="hud-left">
            <a href="/" class="btn-icon" title="Voltar à Biblioteca">
                <svg viewBox="0 0 24 24"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>
            </a>
            <button id="btn-toc" class="btn-icon" title="Índice (T)">
                <svg viewBox="0 0 24 24"><path d="M4 6h16M4 12h16M4 18h7"/></svg>
            </button>
            <span id="sync-status" class="sync-badge">
                <svg viewBox="0 0 24 24" width="12" height="12"><path d="M20 6L9 17l-5-5"/></svg>
                <span>A sincronizar</span>
            </span>
        </div>

        <div class="hud-center">
            <span class="hud-book-title">{{ title }}</span>
            <span id="hud-chapter" class="hud-chapter-title">A carregar capítulo...</span>
        </div>

        <div class="hud-right">
            <button id="btn-settings" class="btn-icon" title="Definições de Leitura (S)">
                <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
            </button>
            <button id="btn-fullscreen" class="btn-icon" title="Ecrã Completo (F)">
                <svg viewBox="0 0 24 24"><path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"/></svg>
            </button>
        </div>
    </header>

    <!-- Table of Contents Drawer -->
    <aside id="toc-drawer">
        <div class="drawer-header">
            <h3>Índice do Livro</h3>
            <button id="btn-close-toc" class="btn-icon">
                <svg viewBox="0 0 24 24"><path d="M18 6L6 18M6 6l12 12"/></svg>
            </button>
        </div>
        <div class="toc-search">
            <input type="text" id="toc-filter" placeholder="Pesquisar capítulos...">
        </div>
        <ul id="toc-list" class="toc-list"></ul>
    </aside>

    <!-- Settings Popover -->
    <div id="settings-popover">
        <div class="settings-section">
            <div class="settings-label">Tema de Leitura</div>
            <div class="theme-buttons">
                <button class="theme-btn active" data-theme-val="dark">Dark</button>
                <button class="theme-btn" data-theme-val="oled">OLED</button>
                <button class="theme-btn" data-theme-val="sepia">Sepia</button>
                <button class="theme-btn" data-theme-val="light">Light</button>
            </div>
        </div>

        <div class="settings-section">
            <div class="settings-label">Tipo de Letra</div>
            <div class="font-btn-row">
                <button class="theme-btn active" data-font-val="sans" style="flex:1;">Sans</button>
                <button class="theme-btn" data-font-val="serif" style="flex:1;">Serif</button>
                <button class="theme-btn" data-font-val="dyslexic" style="flex:1;">Dyslexic</button>
            </div>
        </div>

        <div class="settings-section">
            <div class="settings-label">Tamanho de Letra</div>
            <div class="font-size-ctrl">
                <button id="font-decrease" class="ctrl-btn">A-</button>
                <span id="font-size-label" style="font-size: 13px; font-weight: 600;">100%</span>
                <button id="font-increase" class="ctrl-btn">A+</button>
            </div>
        </div>

        <div class="settings-section" style="margin-bottom: 0;">
            <div class="settings-label">Disposição de Página</div>
            <div class="font-btn-row">
                <button class="theme-btn active" id="btn-spread-auto" style="flex:1;">Auto</button>
                <button class="theme-btn" id="btn-spread-single" style="flex:1;">Uma Coluna</button>
            </div>
        </div>
    </div>

    <!-- Viewer & Tap Navigation Zones -->
    <main id="viewer-container">
        <div id="viewer"></div>
        <div class="tap-zone tap-prev" id="tap-prev" title="Página Anterior (Seta Esquerda)">
            <div class="tap-btn">
                <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><path d="M15 18l-6-6 6-6"/></svg>
            </div>
        </div>
        <div class="tap-zone tap-next" id="tap-next" title="Página Seguinte (Seta Direita / Espaço)">
            <div class="tap-btn">
                <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 18l6-6-6-6"/></svg>
            </div>
        </div>
    </main>

    <!-- Bottom Floating Progress HUD -->
    <footer id="bottom-hud">
        <span id="reading-chapter-bottom" style="opacity: 0.7; max-width: 25%; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">Início</span>
        <div class="progress-bar-wrap" id="progress-bar-track">
            <div class="progress-bar-fill" id="progress-bar-fill"></div>
        </div>
        <span id="reading-percentage" style="font-weight: 600; min-width: 48px; text-align: right;">0%</span>
    </footer>

    <script>
        const BOOK_ID = "{{ bookID }}";
        const EPUB_URL = "{{ epubURL }}";
        const INITIAL_BOOKMARK = "{{ bookmark }}";

        let book = null;
        let rendition = null;
        let currentFontSize = parseInt(localStorage.getItem("cwa_reader_font_size") || "100", 10);
        let currentTheme = localStorage.getItem("cwa_reader_theme") || "dark";
        let currentFont = localStorage.getItem("cwa_reader_font") || "sans";
        let syncTimeout = null;
        let tocData = [];

        // Apply saved theme
        applyTheme(currentTheme);

        function hideLoading() {
            const loader = document.getElementById("loading");
            if (loader) {
                loader.style.opacity = "0";
                loader.style.transition = "opacity 0.25s ease";
                setTimeout(() => { loader.style.display = "none"; }, 260);
            }
        }

        function showLoadingError(msg) {
            const loader = document.getElementById("loading");
            if (loader) {
                loader.innerHTML = '<div style="text-align:center; padding:32px; max-width:440px; background:rgba(255,255,255,0.05); border:1px solid rgba(255,255,255,0.12); border-radius:16px; backdrop-filter:blur(16px); box-shadow:0 20px 40px rgba(0,0,0,0.5);">' +
                    '<div style="font-size:36px; margin-bottom:12px;">\u26a0\ufe0f</div>' +
                    '<h3 style="color:#ef4444; margin:0 0 8px; font-size:18px; font-weight:600;">Falha ao abrir livro</h3>' +
                    '<p style="opacity:0.75; font-size:13px; line-height:1.5; margin:0 0 20px;">' + (msg || "Ficheiro inacess\u00edvel") + '</p>' +
                    '<a href="/" style="display:inline-block; padding:10px 22px; background:linear-gradient(135deg, #6366f1, #4f46e5); color:#fff; border-radius:8px; text-decoration:none; font-size:13px; font-weight:600;">Voltar \u00e0 Biblioteca</a>' +
                    '</div>';
            }
        }

        // Theme and Style rules
        function getThemeRules(theme) {
            if (theme === "dark") return { body: { background: "#121214 !important", color: "#e4e4e7 !important" } };
            if (theme === "oled") return { body: { background: "#000000 !important", color: "#d1d5db !important" } };
            if (theme === "sepia") return { body: { background: "#f4ecd8 !important", color: "#423528 !important" } };
            return { body: { background: "#ffffff !important", color: "#18181b !important" } };
        }

        // Initialize ePub by fetching binary ArrayBuffer directly into memory
        fetch(EPUB_URL)
            .then(res => {
                if (!res.ok) throw new Error("Erro HTTP " + res.status + " ao carregar ficheiro EPUB.");
                return res.arrayBuffer();
            })
            .then(buffer => {
                if (!buffer || buffer.byteLength === 0) {
                    throw new Error("O ficheiro do livro está vazio ou inacessível.");
                }
                book = ePub(buffer);
                rendition = book.renderTo("viewer", {
                    width: "100%",
                    height: "100%",
                    spread: window.innerWidth > 900 ? "auto" : "none",
                    flow: "paginated"
                });

                rendition.themes.register("dark", getThemeRules("dark"));
                rendition.themes.register("oled", getThemeRules("oled"));
                rendition.themes.register("sepia", getThemeRules("sepia"));
                rendition.themes.register("light", getThemeRules("light"));
                applyTheme(currentTheme);
                applyFont(currentFont);
                updateFontSize();

                // Load Table of Contents
                book.loaded.navigation.then(nav => {
                    tocData = nav.toc || [];
                    renderTOC(tocData);
                }).catch(e => console.warn("TOC loading error:", e));

                // Relocated handler
                rendition.on("relocated", onRelocated);

                // Initial display: Check saved progress first
                return fetch("/api/reader/progress/" + BOOK_ID)
                    .then(r => r.json())
                    .then(data => {
                        if (data && data.progress && data.progress.length > 5) {
                            return rendition.display(data.progress);
                        } else if (INITIAL_BOOKMARK) {
                            return rendition.display(INITIAL_BOOKMARK);
                        }
                        return rendition.display();
                    })
                    .catch(() => {
                        if (INITIAL_BOOKMARK) return rendition.display(INITIAL_BOOKMARK);
                        return rendition.display();
                    });
            })
            .then(() => {
                hideLoading();
                // Asynchronously generate locations in background after UI is visible
                setTimeout(() => {
                    if (book && book.locations) {
                        book.locations.generate(1024).catch(() => {});
                    }
                }, 800);
            })
            .catch(err => {
                console.error("Reader init error:", err);
                showLoadingError(err.message || "Erro desconhecido ao renderizar EPUB.");
            });

        // Relocated event callback
        function onRelocated(location) {
            if (!location || !location.start) return;
            const cfi = location.start.cfi;
            let pct = 0;
            if (book && book.locations && book.locations.total > 0) {
                pct = Math.round(book.locations.percentageFromCfi(cfi) * 100);
            }

            // Update Progress HUD
            const bar = document.getElementById("progress-bar-fill");
            const pctLabel = document.getElementById("reading-percentage");
            if (bar) bar.style.width = pct + "%";
            if (pctLabel) pctLabel.textContent = pct + "%";

            // Update Chapter Title
            if (location.start.href) {
                const found = findChapterByHref(location.start.href, tocData);
                if (found) {
                    const ch = document.getElementById("hud-chapter");
                    const chB = document.getElementById("reading-chapter-bottom");
                    if (ch) ch.textContent = found.label.trim();
                    if (chB) chB.textContent = found.label.trim();
                    highlightActiveTOC(found.href);
                }
            }

            // Debounced Save
            saveProgressDebounced(cfi, pct);
        }

        function saveProgressDebounced(cfi, pct) {
            updateSyncBadge(false, "A guardar...");
            clearTimeout(syncTimeout);
            syncTimeout = setTimeout(() => {
                fetch("/api/reader/progress/" + BOOK_ID, {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        progress: cfi,
                        percentage: pct,
                        device: "Web Reader"
                    })
                })
                .then(r => r.json())
                .then(() => {
                    updateSyncBadge(true, "Sincronizado");
                })
                .catch(() => {
                    updateSyncBadge(false, "Erro ao guardar");
                });
            }, 600);
        }

        function updateSyncBadge(saved, text) {
            const badge = document.getElementById("sync-status");
            badge.className = "sync-badge" + (saved ? " saved" : "");
            badge.querySelector("span").textContent = text;
        }

        // TOC Helpers
        function renderTOC(items) {
            const list = document.getElementById("toc-list");
            list.innerHTML = "";
            items.forEach(item => {
                const li = document.createElement("li");
                li.className = "toc-item";
                li.textContent = item.label.trim();
                li.dataset.href = item.href;
                li.addEventListener("click", () => {
                    rendition.display(item.href);
                    closeDrawers();
                });
                list.appendChild(li);
                if (item.subitems && item.subitems.length) {
                    item.subitems.forEach(sub => {
                        const subLi = document.createElement("li");
                        subLi.className = "toc-item";
                        subLi.style.paddingLeft = "32px";
                        subLi.textContent = sub.label.trim();
                        subLi.dataset.href = sub.href;
                        subLi.addEventListener("click", () => {
                            rendition.display(sub.href);
                            closeDrawers();
                        });
                        list.appendChild(subLi);
                    });
                }
            });
        }

        function findChapterByHref(href, items) {
            for (let item of items) {
                if (href.includes(item.href) || item.href.includes(href)) return item;
                if (item.subitems && item.subitems.length) {
                    const sub = findChapterByHref(href, item.subitems);
                    if (sub) return sub;
                }
            }
            return null;
        }

        function highlightActiveTOC(href) {
            document.querySelectorAll(".toc-item").forEach(el => {
                if (el.dataset.href && (href.includes(el.dataset.href) || el.dataset.href.includes(href))) {
                    el.classList.add("active");
                } else {
                    el.classList.remove("active");
                }
            });
        }

        // TOC Filter
        document.getElementById("toc-filter").addEventListener("input", function(e) {
            const term = e.target.value.toLowerCase();
            document.querySelectorAll(".toc-item").forEach(el => {
                el.style.display = el.textContent.toLowerCase().includes(term) ? "" : "none";
            });
        });

        // Theme management
        function applyTheme(theme) {
            document.body.setAttribute("data-theme", theme);
            currentTheme = theme;
            localStorage.setItem("cwa_reader_theme", theme);
            document.querySelectorAll(".theme-btn[data-theme-val]").forEach(b => {
                b.classList.toggle("active", b.dataset.themeVal === theme);
            });
            if (rendition) rendition.themes.select(theme);
        }

        document.querySelectorAll(".theme-btn[data-theme-val]").forEach(btn => {
            btn.addEventListener("click", () => applyTheme(btn.dataset.themeVal));
        });

        // Font Family
        function applyFont(font) {
            currentFont = font;
            localStorage.setItem("cwa_reader_font", font);
            document.querySelectorAll(".theme-btn[data-font-val]").forEach(b => {
                b.classList.toggle("active", b.dataset.fontVal === font);
            });
            let fontFamily = "-apple-system, BlinkMacSystemFont, Segoe UI, Roboto, sans-serif";
            if (font === "serif") fontFamily = "Merriweather, Georgia, Cambria, serif";
            if (font === "dyslexic") fontFamily = "OpenDyslexic, Comic Sans MS, sans-serif";
            if (rendition) rendition.themes.font(fontFamily);
        }

        document.querySelectorAll(".theme-btn[data-font-val]").forEach(btn => {
            btn.addEventListener("click", () => applyFont(btn.dataset.fontVal));
        });

        // Font Size
        document.getElementById("font-increase").addEventListener("click", () => {
            if (currentFontSize < 250) {
                currentFontSize += 10;
                updateFontSize();
            }
        });
        document.getElementById("font-decrease").addEventListener("click", () => {
            if (currentFontSize > 70) {
                currentFontSize -= 10;
                updateFontSize();
            }
        });

        function updateFontSize() {
            document.getElementById("font-size-label").textContent = currentFontSize + "%";
            localStorage.setItem("cwa_reader_font_size", currentFontSize);
            if (rendition) rendition.themes.fontSize(currentFontSize + "%");
        }
        updateFontSize();

        // Spread Toggle
        document.getElementById("btn-spread-auto").addEventListener("click", function() {
            this.classList.add("active");
            document.getElementById("btn-spread-single").classList.remove("active");
            rendition.spread("auto");
        });
        document.getElementById("btn-spread-single").addEventListener("click", function() {
            this.classList.add("active");
            document.getElementById("btn-spread-auto").classList.remove("active");
            rendition.spread("none");
        });

        // Navigation
        document.getElementById("tap-prev").addEventListener("click", () => rendition.prev());
        document.getElementById("tap-next").addEventListener("click", () => rendition.next());

        // Keyboard Shortcuts
        document.addEventListener("keydown", e => {
            if (e.target.tagName === "INPUT") return;
            if (e.key === "ArrowLeft" || e.key === "j" || e.key === "PageUp") { rendition.prev(); }
            else if (e.key === "ArrowRight" || e.key === "k" || e.key === " " || e.key === "PageDown") { rendition.next(); }
            else if (e.key === "t" || e.key === "T") { toggleTOC(); }
            else if (e.key === "s" || e.key === "S") { toggleSettings(); }
            else if (e.key === "f" || e.key === "F") { toggleFullscreen(); }
            else if (e.key === "Escape") { closeDrawers(); }
        });

        // Drawer & Popover toggles
        const drawer = document.getElementById("toc-drawer");
        const popover = document.getElementById("settings-popover");
        const backdrop = document.getElementById("backdrop");

        function toggleTOC() {
            popover.classList.remove("open");
            drawer.classList.toggle("open");
            backdrop.classList.toggle("open", drawer.classList.contains("open"));
        }

        function toggleSettings() {
            drawer.classList.remove("open");
            popover.classList.toggle("open");
            backdrop.classList.toggle("open", popover.classList.contains("open"));
        }

        function closeDrawers() {
            drawer.classList.remove("open");
            popover.classList.remove("open");
            backdrop.classList.remove("open");
        }

        document.getElementById("btn-toc").addEventListener("click", toggleTOC);
        document.getElementById("btn-close-toc").addEventListener("click", closeDrawers);
        document.getElementById("btn-settings").addEventListener("click", toggleSettings);
        backdrop.addEventListener("click", closeDrawers);

        // Fullscreen
        function toggleFullscreen() {
            if (!document.fullscreenElement) {
                document.documentElement.requestFullscreen().catch(() => {});
            } else {
                document.exitFullscreen().catch(() => {});
            }
        }
        document.getElementById("btn-fullscreen").addEventListener("click", toggleFullscreen);
    </script>
</body>
</html>`


