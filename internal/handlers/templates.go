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
                <a href="/cwa-sources" class="cwa-op-card">
                    <div>
                        <div class="op-title"><span class="glyphicon glyphicon-cloud-download" style="color: #38bdf8; margin-right: 8px;"></span> Book Sources &amp; Acquisition Hub</div>
                        <div class="op-desc">Download books directly from Open Library, Gutenberg, Anna's Archive &amp; Z-Library</div>
                    </div>
                    <span class="glyphicon glyphicon-chevron-right" style="color: #64748b; font-size: 12px;"></span>
                </a>
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
<html lang="pt">
<head>
    <title>{{ title }} — CWA NextGen Reader</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no, viewport-fit=cover">
    <link rel="apple-touch-icon" sizes="140x140" href="/static/favicon.ico">
    <link rel="shortcut icon" href="/static/favicon.ico">
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&family=JetBrains+Mono:wght@400;500&family=Literata:ital,opsz,wght@0,7..72,400;0,7..72,500;0,7..72,600;1,7..72,400&family=Merriweather:ital,wght@0,300;0,400;0,700;1,300&display=swap" rel="stylesheet">
    <script src="/static/js/libs/jquery.min.js"></script>
    <script src="/static/js/compress/jszip_epub.min.js"></script>
    <script src="/static/js/libs/epub.min.js"></script>
    <style>
        :root {
            --bg-color: #0b0f19;
            --text-color: #e2e8f0;
            --hud-bg: rgba(15, 23, 42, 0.94);
            --border-color: rgba(255, 255, 255, 0.08);
            --border-hover: rgba(255, 255, 255, 0.16);
            --accent-color: #38bdf8;
            --accent-subtle: rgba(56, 189, 248, 0.12);
            --sidebar-bg: #0f172a;
            --surface-hover: rgba(255, 255, 255, 0.06);
            --page-width: 860px;
            --font-family: 'Literata', Georgia, Cambria, serif;
            --safe-top: env(safe-area-inset-top, 0px);
            --safe-bottom: env(safe-area-inset-bottom, 0px);
            --safe-left: env(safe-area-inset-left, 0px);
            --safe-right: env(safe-area-inset-right, 0px);
        }

        body[data-theme="dark"] {
            --bg-color: #0b0f19;
            --text-color: #e2e8f0;
            --hud-bg: rgba(15, 23, 42, 0.94);
            --border-color: rgba(255, 255, 255, 0.08);
            --border-hover: rgba(255, 255, 255, 0.16);
            --accent-color: #38bdf8;
            --accent-subtle: rgba(56, 189, 248, 0.12);
            --sidebar-bg: #0f172a;
            --surface-hover: rgba(255, 255, 255, 0.06);
        }

        body[data-theme="oled"] {
            --bg-color: #000000;
            --text-color: #d1d5db;
            --hud-bg: rgba(9, 9, 11, 0.96);
            --border-color: rgba(255, 255, 255, 0.12);
            --border-hover: rgba(255, 255, 255, 0.22);
            --accent-color: #38bdf8;
            --accent-subtle: rgba(56, 189, 248, 0.12);
            --sidebar-bg: #09090b;
            --surface-hover: rgba(255, 255, 255, 0.08);
        }

        body[data-theme="sepia"] {
            --bg-color: #fbf0d9;
            --text-color: #3e3020;
            --hud-bg: rgba(246, 235, 210, 0.95);
            --border-color: rgba(62, 48, 32, 0.12);
            --border-hover: rgba(62, 48, 32, 0.24);
            --accent-color: #b45309;
            --accent-subtle: rgba(180, 83, 9, 0.12);
            --sidebar-bg: #f4e3be;
            --surface-hover: rgba(62, 48, 32, 0.06);
        }

        body[data-theme="light"] {
            --bg-color: #f8fafc;
            --text-color: #0f172a;
            --hud-bg: rgba(255, 255, 255, 0.95);
            --border-color: rgba(0, 0, 0, 0.08);
            --border-hover: rgba(0, 0, 0, 0.16);
            --accent-color: #0284c7;
            --accent-subtle: rgba(2, 132, 199, 0.1);
            --sidebar-bg: #ffffff;
            --surface-hover: rgba(0, 0, 0, 0.05);
        }

        * { box-sizing: border-box; }
        html, body {
            margin: 0; padding: 0; width: 100%; height: 100%;
            overflow: hidden; background: var(--bg-color); color: var(--text-color);
            font-family: var(--font-family); user-select: none; -webkit-user-select: none;
            transition: background-color 0.25s ease, color 0.25s ease;
        }

        /* Top HUD */
        #top-hud {
            position: fixed; top: 0; left: 0; right: 0;
            padding-top: var(--safe-top);
            height: calc(52px + var(--safe-top));
            background: var(--hud-bg); backdrop-filter: blur(16px); -webkit-backdrop-filter: blur(16px);
            border-bottom: 1px solid var(--border-color); display: flex; align-items: center;
            justify-content: space-between; padding-left: max(16px, var(--safe-left)); padding-right: max(16px, var(--safe-right));
            z-index: 60; transition: transform 0.24s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.24s ease;
        }
        #top-hud.is-hidden { transform: translateY(-100%); opacity: 0; pointer-events: none; }

        .hud-left, .hud-right { display: flex; align-items: center; gap: 6px; }
        .hud-center {
            position: absolute; left: 50%; transform: translateX(-50%);
            display: flex; flex-direction: column; align-items: center;
            max-width: 44%; pointer-events: none; text-align: center;
        }
        .hud-book-title {
            font-size: 13px; font-weight: 600; white-space: nowrap; overflow: hidden;
            text-overflow: ellipsis; max-width: 100%; letter-spacing: -0.01em;
        }
        .hud-chapter-title {
            font-size: 11px; opacity: 0.65; white-space: nowrap; overflow: hidden;
            text-overflow: ellipsis; max-width: 100%; margin-top: 1px;
        }

        .btn-icon {
            background: transparent; border: 1px solid transparent; color: var(--text-color);
            border-radius: 8px; width: 34px; height: 34px; display: inline-flex; align-items: center;
            justify-content: center; cursor: pointer; transition: all 0.15s ease; text-decoration: none;
            padding: 0; flex-shrink: 0;
        }
        .btn-icon:hover { background: var(--surface-hover); border-color: var(--border-color); }
        .btn-icon.active { background: var(--accent-subtle); border-color: var(--accent-color); color: var(--accent-color); }
        .btn-icon svg { width: 17px; height: 17px; fill: none; stroke: currentColor; stroke-width: 2; stroke-linecap: round; stroke-linejoin: round; }

        .sync-badge {
            font-size: 11px; padding: 2px 7px; border-radius: 9999px; background: var(--accent-subtle);
            color: var(--accent-color); font-weight: 500; display: inline-flex; align-items: center; gap: 4px;
            transition: all 0.2s ease;
        }
        .sync-badge.saved { background: rgba(34, 197, 94, 0.14); color: #22c55e; }

        /* Viewer container & Layout */
        #viewer-container {
            position: absolute; inset: 0;
            display: flex; align-items: center; justify-content: center;
            overflow: hidden;
        }
        #viewer-wrapper {
            width: 100%; height: 100%;
            max-width: var(--page-width);
            margin: 0 auto; position: relative;
            transition: max-width 0.2s ease;
        }
        #viewer {
            width: 100%; height: 100%;
        }

        /* Continuous scroll mode adjustments */
        body.mode-scroll #viewer-container {
            overflow-y: auto; overflow-x: hidden; -webkit-overflow-scrolling: touch;
        }
        body.mode-scroll #viewer-wrapper {
            height: auto; min-height: 100%;
        }

        /* Floating Nav Chevrons (Desktop) */
        .edge-nav {
            position: fixed; top: 50%; transform: translateY(-50%); width: 48px; height: 72px;
            display: flex; align-items: center; justify-content: center; z-index: 40;
            cursor: pointer; opacity: 0; transition: opacity 0.2s ease, transform 0.15s ease;
            color: var(--text-color); border: none; background: transparent; padding: 0;
        }
        .edge-nav:hover { opacity: 0.85; transform: translateY(-50%) scale(1.05); }
        .edge-nav-inner {
            width: 38px; height: 38px; border-radius: 50%;
            background: var(--hud-bg); border: 1px solid var(--border-color);
            display: flex; align-items: center; justify-content: center;
            box-shadow: 0 4px 16px rgba(0,0,0,0.3);
        }
        .edge-nav-prev { left: 16px; }
        .edge-nav-next { right: 16px; }
        @media (max-width: 768px) {
            .edge-nav { display: none !important; }
        }

        /* Bottom Floating Progress HUD */
        #bottom-hud {
            position: fixed; bottom: 0; left: 0; right: 0;
            padding-bottom: var(--safe-bottom);
            height: calc(48px + var(--safe-bottom));
            background: var(--hud-bg); backdrop-filter: blur(16px); -webkit-backdrop-filter: blur(16px);
            border-top: 1px solid var(--border-color); display: flex; align-items: center;
            justify-content: space-between; padding-left: max(18px, var(--safe-left)); padding-right: max(18px, var(--safe-right));
            z-index: 60; font-size: 12px; transition: transform 0.24s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.24s ease;
        }
        #bottom-hud.is-hidden { transform: translateY(100%); opacity: 0; pointer-events: none; }

        .progress-section {
            flex: 1; max-width: 560px; display: flex; align-items: center; gap: 12px; margin: 0 16px;
        }
        .progress-track {
            flex: 1; height: 6px; background: rgba(255, 255, 255, 0.12);
            border-radius: 9999px; position: relative; cursor: pointer; padding: 6px 0;
            background-clip: content-box;
        }
        .progress-fill {
            height: 6px; background: var(--accent-color); width: 0%;
            border-radius: 9999px; transition: width 0.15s ease; pointer-events: none;
        }
        .progress-thumb {
            position: absolute; top: 3px; width: 12px; height: 12px; border-radius: 50%;
            background: #ffffff; box-shadow: 0 1px 4px rgba(0,0,0,0.5); transform: translateX(-50%);
            display: none; pointer-events: none;
        }
        .progress-track:hover .progress-thumb { display: block; }

        .reading-metric {
            font-size: 11px; font-variant-numeric: tabular-nums; opacity: 0.75; white-space: nowrap;
        }

        /* Unified Drawer (Table of Contents, Search, Bookmarks) */
        #reader-drawer {
            position: fixed; top: 0; bottom: 0; left: 0; width: 380px; max-width: 90vw;
            background: var(--sidebar-bg); border-right: 1px solid var(--border-color);
            z-index: 100; transform: translateX(-100%); transition: transform 0.25s cubic-bezier(0.16, 1, 0.3, 1);
            display: flex; flex-direction: column; box-shadow: 6px 0 28px rgba(0,0,0,0.45);
        }
        #reader-drawer.open { transform: translateX(0); }

        .drawer-header {
            padding: 12px 16px; border-bottom: 1px solid var(--border-color);
            display: flex; align-items: center; justify-content: space-between; gap: 8px;
            padding-top: max(12px, var(--safe-top));
        }
        .drawer-tabs {
            display: flex; align-items: center; gap: 4px; background: rgba(255, 255, 255, 0.05);
            padding: 3px; border-radius: 8px; flex: 1;
        }
        .drawer-tab {
            flex: 1; border: none; background: transparent; color: var(--text-color);
            font-size: 12px; font-weight: 500; padding: 6px 8px; border-radius: 6px;
            cursor: pointer; text-align: center; transition: all 0.15s; opacity: 0.7;
        }
        .drawer-tab.active { background: var(--hud-bg); opacity: 1; font-weight: 600; color: var(--accent-color); }

        .drawer-content { flex: 1; overflow-y: auto; display: none; padding: 12px 0; }
        .drawer-content.active { display: block; }

        .drawer-search-box {
            padding: 8px 16px 12px; border-bottom: 1px solid var(--border-color);
            display: flex; gap: 8px; align-items: center;
        }
        .drawer-search-input {
            flex: 1; background: rgba(255, 255, 255, 0.06); border: 1px solid var(--border-color);
            border-radius: 8px; padding: 7px 12px; color: var(--text-color); font-size: 13px; outline: none;
            transition: border-color 0.15s;
        }
        .drawer-search-input:focus { border-color: var(--accent-color); }

        .toc-list, .bookmarks-list, .search-results-list {
            list-style: none; margin: 0; padding: 0;
        }
        .toc-item, .bookmark-item, .search-item {
            padding: 10px 18px; font-size: 13px; cursor: pointer; border-left: 3px solid transparent;
            transition: background 0.15s, border-color 0.15s;
        }
        .toc-item:hover, .bookmark-item:hover, .search-item:hover { background: var(--surface-hover); }
        .toc-item.active { border-color: var(--accent-color); color: var(--accent-color); font-weight: 600; background: var(--accent-subtle); }

        .search-item-snippet {
            font-size: 12px; opacity: 0.8; margin-top: 4px; line-height: 1.4;
        }
        .search-item-snippet mark {
            background: rgba(56, 189, 248, 0.3); color: inherit; padding: 1px 3px; border-radius: 2px;
        }
        .bookmark-actions {
            display: flex; align-items: center; justify-content: space-between; font-size: 11px; opacity: 0.6; margin-top: 4px;
        }

        .drawer-empty-state {
            padding: 36px 20px; text-align: center; opacity: 0.6; font-size: 13px;
        }

        /* Settings Popover (Aa) */
        #settings-popover {
            position: fixed; top: calc(56px + var(--safe-top)); right: max(16px, var(--safe-right)); width: 330px;
            background: var(--sidebar-bg); border: 1px solid var(--border-color);
            border-radius: 14px; padding: 18px; z-index: 100; display: none;
            box-shadow: 0 16px 40px rgba(0,0,0,0.45); animation: popoverFadeIn 0.15s ease;
        }
        #settings-popover.open { display: block; }
        @keyframes popoverFadeIn { from { opacity: 0; transform: translateY(-8px); } to { opacity: 1; transform: translateY(0); } }

        .settings-row { margin-bottom: 16px; }
        .settings-row:last-child { margin-bottom: 0; }
        .settings-title {
            font-size: 11px; text-transform: uppercase; letter-spacing: 0.06em; opacity: 0.65;
            margin-bottom: 8px; font-weight: 600;
        }
        .pill-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 6px; }
        .pill-btn {
            padding: 7px 4px; border-radius: 7px; border: 1px solid var(--border-color);
            cursor: pointer; font-size: 11px; font-weight: 500; text-align: center;
            background: rgba(255, 255, 255, 0.04); color: var(--text-color); transition: all 0.15s;
        }
        .pill-btn:hover { background: var(--surface-hover); border-color: var(--border-hover); }
        .pill-btn.active { border-color: var(--accent-color); background: var(--accent-subtle); color: var(--accent-color); font-weight: 600; }

        .stepper-ctrl {
            display: flex; align-items: center; justify-content: space-between;
            background: rgba(255, 255, 255, 0.04); border: 1px solid var(--border-color);
            border-radius: 8px; padding: 3px 8px;
        }
        .stepper-btn {
            background: transparent; border: none; color: var(--text-color); font-size: 14px;
            font-weight: 600; cursor: pointer; padding: 4px 10px; border-radius: 6px;
        }
        .stepper-btn:hover { background: var(--surface-hover); }

        /* Backdrop overlay */
        #backdrop {
            position: fixed; inset: 0; background: rgba(0, 0, 0, 0.55); z-index: 80;
            display: none; backdrop-filter: blur(2px); -webkit-backdrop-filter: blur(2px);
        }
        #backdrop.open { display: block; }

        /* Loading / Error Screen */
        #loading-screen {
            position: fixed; inset: 0; display: flex; flex-direction: column;
            align-items: center; justify-content: center; background: var(--bg-color); z-index: 200;
            transition: opacity 0.25s ease;
        }
        .spinner-ring {
            width: 38px; height: 38px; border: 3px solid rgba(56, 189, 248, 0.18);
            border-top-color: var(--accent-color); border-radius: 50%;
            animation: spin 0.75s linear infinite;
        }
        @keyframes spin { to { transform: rotate(360deg); } }
    </style>
</head>
<body data-theme="dark">
    <!-- Loading Screen -->
    <div id="loading-screen">
        <div class="spinner-ring"></div>
        <p id="loading-text" style="margin-top: 16px; font-size: 13px; opacity: 0.8;">A carregar livro...</p>
    </div>

    <!-- Backdrop -->
    <div id="backdrop"></div>

    <!-- Top HUD -->
    <header id="top-hud">
        <div class="hud-left">
            <a href="/" class="btn-icon" title="Voltar à Biblioteca">
                <svg viewBox="0 0 24 24"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>
            </a>
            <button id="btn-open-toc" class="btn-icon" title="Índice & Pesquisa (T)">
                <svg viewBox="0 0 24 24"><path d="M4 6h16M4 12h16M4 18h7"/></svg>
            </button>
            <button id="btn-bookmark" class="btn-icon" title="Guardar Marcador (B)">
                <svg viewBox="0 0 24 24"><path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z"/></svg>
            </button>
            <span id="sync-status" class="sync-badge" style="margin-left: 4px;">
                <svg viewBox="0 0 24 24" width="11" height="11"><path d="M20 6L9 17l-5-5"/></svg>
                <span>Pronto</span>
            </span>
        </div>

        <div class="hud-center">
            <span class="hud-book-title">{{ title }}</span>
            <span id="hud-chapter" class="hud-chapter-title">A inicializar...</span>
        </div>

        <div class="hud-right">
            <button id="btn-toggle-flow" class="btn-icon" title="Alternar Modo Paginado / Scroll Contínuo (M)">
                <svg id="icon-flow-paginated" viewBox="0 0 24 24"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/></svg>
                <svg id="icon-flow-scroll" viewBox="0 0 24 24" style="display:none;"><path d="M12 3v18M8 7l4-4 4 4M8 17l4 4 4-4"/></svg>
            </button>
            <button id="btn-settings" class="btn-icon" title="Tipografia e Aspeto (Aa)">
                <svg viewBox="0 0 24 24"><path d="M4 20h16M14 4l-4 12m-2-5h5"/></svg>
            </button>
            <button id="btn-fullscreen" class="btn-icon" title="Ecrã Completo (F)">
                <svg viewBox="0 0 24 24"><path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"/></svg>
            </button>
        </div>
    </header>

    <!-- Unified Reader Drawer (TOC, Search, Bookmarks) -->
    <aside id="reader-drawer">
        <div class="drawer-header">
            <div class="drawer-tabs">
                <button class="drawer-tab active" data-tab="toc">Índice</button>
                <button class="drawer-tab" data-tab="search">Pesquisa</button>
                <button class="drawer-tab" data-tab="bookmarks">Marcadores</button>
            </div>
            <button id="btn-close-drawer" class="btn-icon" title="Fechar (Esc)">
                <svg viewBox="0 0 24 24"><path d="M18 6L6 18M6 6l12 12"/></svg>
            </button>
        </div>

        <!-- Tab 1: TOC -->
        <div id="tab-content-toc" class="drawer-content active">
            <div class="drawer-search-box">
                <input type="text" id="toc-filter-input" class="drawer-search-input" placeholder="Filtrar capítulos...">
            </div>
            <ul id="toc-list" class="toc-list"></ul>
        </div>

        <!-- Tab 2: In-Book Search -->
        <div id="tab-content-search" class="drawer-content">
            <div class="drawer-search-box">
                <input type="text" id="book-search-input" class="drawer-search-input" placeholder="Pesquisar texto no livro...">
                <button id="btn-run-search" class="btn-icon active" style="width:30px; height:30px;" title="Pesquisar">
                    <svg viewBox="0 0 24 24"><circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/></svg>
                </button>
            </div>
            <div id="search-status-bar" style="padding: 6px 16px; font-size: 11px; opacity: 0.6; display: none;"></div>
            <ul id="search-results-list" class="search-results-list"></ul>
        </div>

        <!-- Tab 3: Bookmarks -->
        <div id="tab-content-bookmarks" class="drawer-content">
            <div style="padding: 8px 16px 12px; border-bottom: 1px solid var(--border-color);">
                <button id="btn-add-bookmark-drawer" class="pill-btn active" style="width:100%; padding: 8px; font-size: 12px;">+ Guardar Posição Atual</button>
            </div>
            <ul id="bookmarks-list" class="bookmarks-list"></ul>
        </div>
    </aside>

    <!-- Settings Popover (Aa) -->
    <div id="settings-popover">
        <!-- Theme row -->
        <div class="settings-row">
            <div class="settings-title">Tema de Leitura</div>
            <div class="pill-grid">
                <button class="pill-btn active" data-theme-btn="dark">Dark</button>
                <button class="pill-btn" data-theme-btn="oled">OLED</button>
                <button class="pill-btn" data-theme-btn="sepia">Sépia</button>
                <button class="pill-btn" data-theme-btn="light">Light</button>
            </div>
        </div>

        <!-- Font Family -->
        <div class="settings-row">
            <div class="settings-title">Tipo de Letra</div>
            <div class="pill-grid" style="grid-template-columns: repeat(3, 1fr);">
                <button class="pill-btn active" data-font-btn="literata">Literata</button>
                <button class="pill-btn" data-font-btn="inter">Inter Sans</button>
                <button class="pill-btn" data-font-btn="merriweather">Merriweather</button>
            </div>
            <div class="pill-grid" style="grid-template-columns: repeat(2, 1fr); margin-top: 6px;">
                <button class="pill-btn" data-font-btn="mono">Mono</button>
                <button class="pill-btn" data-font-btn="dyslexic">OpenDyslexic</button>
            </div>
        </div>

        <!-- Font Size -->
        <div class="settings-row">
            <div class="settings-title">Tamanho da Letra</div>
            <div class="stepper-ctrl">
                <button id="btn-font-dec" class="stepper-btn">A−</button>
                <span id="font-size-display" style="font-size: 12px; font-weight: 600;">100%</span>
                <button id="btn-font-inc" class="stepper-btn">A+</button>
            </div>
        </div>

        <!-- Line Height -->
        <div class="settings-row">
            <div class="settings-title">Espaçamento de Linha</div>
            <div class="pill-grid" style="grid-template-columns: repeat(3, 1fr);">
                <button class="pill-btn" data-lh-btn="1.35">Compacto</button>
                <button class="pill-btn active" data-lh-btn="1.6">Normal</button>
                <button class="pill-btn" data-lh-btn="1.9">Amplo</button>
            </div>
        </div>

        <!-- Page Width / Margins -->
        <div class="settings-row">
            <div class="settings-title">Largura da Página</div>
            <div class="pill-grid">
                <button class="pill-btn" data-width-btn="680px">Estreita</button>
                <button class="pill-btn active" data-width-btn="860px">Normal</button>
                <button class="pill-btn" data-width-btn="1040px">Ampla</button>
                <button class="pill-btn" data-width-btn="100%">Total</button>
            </div>
        </div>

        <!-- Text Alignment -->
        <div class="settings-row">
            <div class="settings-title">Alinhamento</div>
            <div class="pill-grid" style="grid-template-columns: repeat(2, 1fr);">
                <button class="pill-btn active" data-align-btn="justify">Justificado</button>
                <button class="pill-btn" data-align-btn="left">À Esquerda</button>
            </div>
        </div>
    </div>

    <!-- Viewer & Navigation -->
    <main id="viewer-container">
        <div id="viewer-wrapper">
            <div id="viewer"></div>
        </div>
        <!-- Edge Chevrons (Desktop) -->
        <button class="edge-nav edge-nav-prev" id="edge-prev" title="Página Anterior">
            <div class="edge-nav-inner">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M15 18l-6-6 6-6"/></svg>
            </div>
        </button>
        <button class="edge-nav edge-nav-next" id="edge-next" title="Página Seguinte">
            <div class="edge-nav-inner">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M9 18l6-6-6-6"/></svg>
            </div>
        </button>
    </main>

    <!-- Bottom HUD -->
    <footer id="bottom-hud">
        <div style="display: flex; align-items: center; gap: 8px; max-width: 32%;">
            <button id="btn-prev-chapter" class="btn-icon" style="width:28px; height:28px;" title="Capítulo Anterior">
                <svg viewBox="0 0 24 24"><path d="M11 19l-7-7 7-7m8 14l-7-7 7-7"/></svg>
            </button>
            <span id="reading-chapter-bottom" style="opacity: 0.8; font-size: 11px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;">Início</span>
        </div>

        <div class="progress-section">
            <div class="progress-track" id="progress-track" title="Navegar no livro">
                <div class="progress-fill" id="progress-fill"></div>
                <div class="progress-thumb" id="progress-thumb"></div>
            </div>
        </div>

        <div style="display: flex; align-items: center; gap: 8px;">
            <span id="reading-time-left" class="reading-metric">~...</span>
            <span id="reading-percentage" class="reading-metric" style="font-weight: 600; min-width: 36px; text-align: right;">0%</span>
            <button id="btn-next-chapter" class="btn-icon" style="width:28px; height:28px;" title="Próximo Capítulo">
                <svg viewBox="0 0 24 24"><path d="M13 5l7 7-7 7M5 5l7 7-7 7"/></svg>
            </button>
        </div>
    </footer>

    <script>
        const BOOK_ID = "{{ bookID }}";
        const EPUB_URL = "{{ epubURL }}";
        const INITIAL_BOOKMARK = "{{ bookmark }}";

        let book = null;
        let rendition = null;
        let tocData = [];
        let syncTimeout = null;
        let hudTimer = null;
        let hudLocked = false;

        // User Reading Preferences with robust defaults
        let currentTheme = localStorage.getItem("cwa_reader_theme") || "dark";
        let currentFont = localStorage.getItem("cwa_reader_font") || "literata";
        let currentFontSize = parseInt(localStorage.getItem("cwa_reader_font_size") || "100", 10);
        let currentLineHeight = localStorage.getItem("cwa_reader_lh") || "1.6";
        let currentPageWidth = localStorage.getItem("cwa_reader_width") || "860px";
        let currentTextAlign = localStorage.getItem("cwa_reader_align") || "justify";
        let currentFlow = localStorage.getItem("cwa_reader_flow") || "paginated"; // 'paginated' or 'scrolled-doc'

        // Apply theme & width immediately to avoid flash
        document.body.setAttribute("data-theme", currentTheme);
        document.documentElement.style.setProperty("--page-width", currentPageWidth);
        updateFlowIcon();

        // Loading Screen Helpers
        function hideLoading() {
            const loader = document.getElementById("loading-screen");
            if (loader) {
                loader.style.opacity = "0";
                setTimeout(function() { loader.style.display = "none"; }, 260);
            }
        }

        function showLoadingError(msg) {
            const loader = document.getElementById("loading-screen");
            if (loader) {
                loader.innerHTML = '<div style="text-align:center; padding:32px 24px; max-width:440px; background:rgba(255,255,255,0.06); border:1px solid rgba(255,255,255,0.12); border-radius:16px; backdrop-filter:blur(16px);">' +
                    '<div style="font-size:36px; margin-bottom:12px;">\u26a0\ufe0f</div>' +
                    '<h3 style="color:#ef4444; margin:0 0 8px; font-size:18px; font-weight:600;">Falha ao abrir livro</h3>' +
                    '<p style="opacity:0.75; font-size:13px; line-height:1.5; margin:0 0 20px;">' + (msg || "Ficheiro inacess\u00edvel ou formato inv\u00e1lido.") + '</p>' +
                    '<a href="/" style="display:inline-block; padding:10px 22px; background:#0284c7; color:#fff; border-radius:8px; text-decoration:none; font-size:13px; font-weight:600;">Voltar \u00e0 Biblioteca</a>' +
                    '</div>';
            }
        }

        // Font family resolver
        function getFontFamilyString(key) {
            if (key === "inter") return "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif";
            if (key === "merriweather") return "'Merriweather', Georgia, serif";
            if (key === "mono") return "'JetBrains Mono', Consolas, Menlo, monospace";
            if (key === "dyslexic") return "'OpenDyslexic', 'Comic Sans MS', sans-serif";
            return "'Literata', Georgia, Cambria, serif";
        }

        function getThemeBg(t) {
            if (t === "oled") return "#000000";
            if (t === "sepia") return "#fbf0d9";
            if (t === "light") return "#f8fafc";
            return "#0b0f19";
        }

        function getThemeFg(t) {
            if (t === "oled") return "#d1d5db";
            if (t === "sepia") return "#3e3020";
            if (t === "light") return "#0f172a";
            return "#e2e8f0";
        }

        function applyRenditionStyles() {
            if (!rendition) return;
            const bg = getThemeBg(currentTheme);
            const fg = getThemeFg(currentTheme);
            const font = getFontFamilyString(currentFont);

            const rules = {
                "body": {
                    "background-color": bg + " !important",
                    "color": fg + " !important",
                    "font-family": font + " !important",
                    "font-size": currentFontSize + "% !important",
                    "line-height": currentLineHeight + " !important",
                    "text-align": currentTextAlign + " !important",
                    "padding": currentFlow === "scrolled-doc" ? "24px 20px 80px 20px !important" : "0 14px !important",
                    "-webkit-font-smoothing": "antialiased"
                },
                "p": {
                    "line-height": currentLineHeight + " !important",
                    "margin-bottom": "0.85em !important",
                    "text-align": currentTextAlign + " !important"
                },
                "a": {
                    "color": (currentTheme === "sepia" ? "#b45309" : "#38bdf8") + " !important"
                },
                "img": {
                    "max-width": "100% !important",
                    "height": "auto !important",
                    "display": "block !important",
                    "margin": "14px auto !important"
                }
            };

            rendition.themes.default(rules);
            if (rendition.views && rendition.views()) {
                rendition.views().forEach(function(view) {
                    if (view && view.pane) view.pane.render();
                });
            }
        }

        // HUD Visibility Controls
        function showHUD() {
            document.getElementById("top-hud").classList.remove("is-hidden");
            document.getElementById("bottom-hud").classList.remove("is-hidden");
            resetHUDTimer();
        }

        function hideHUD() {
            if (hudLocked) return;
            document.getElementById("top-hud").classList.add("is-hidden");
            document.getElementById("bottom-hud").classList.add("is-hidden");
        }

        function toggleHUD() {
            const isHidden = document.getElementById("top-hud").classList.contains("is-hidden");
            if (isHidden) showHUD();
            else hideHUD();
        }

        function resetHUDTimer() {
            clearTimeout(hudTimer);
            if (!hudLocked) {
                hudTimer = setTimeout(function() {
                    hideHUD();
                }, 4000);
            }
        }

        // Attach touch and click handlers inside epub iframe
        function bindRenditionHooks() {
            if (!rendition) return;

            rendition.hooks.content.register(function(contents) {
                const doc = contents.document;
                if (!doc) return;

                let startX = 0;
                let startY = 0;
                let startTime = 0;

                doc.addEventListener("touchstart", function(e) {
                    if (e.touches.length === 1) {
                        startX = e.touches[0].clientX;
                        startY = e.touches[0].clientY;
                        startTime = Date.now();
                    }
                }, { passive: true });

                doc.addEventListener("touchend", function(e) {
                    if (e.changedTouches.length === 1) {
                        const diffX = e.changedTouches[0].clientX - startX;
                        const diffY = e.changedTouches[0].clientY - startY;
                        const duration = Date.now() - startTime;

                        // Horizontal swipe detection (> 45px, duration < 500ms, mostly horizontal)
                        if (Math.abs(diffX) > 45 && Math.abs(diffX) > Math.abs(diffY) * 1.5 && duration < 500) {
                            if (currentFlow === "paginated") {
                                if (diffX < 0) rendition.next();
                                else rendition.prev();
                                return;
                            }
                        }
                    }
                }, { passive: true });

                doc.addEventListener("click", function(e) {
                    // Ignore clicks on links
                    if (e.target && e.target.closest("a")) return;

                    // If user is selecting text, do not flip or toggle HUD
                    const sel = doc.getSelection();
                    if (sel && sel.toString().trim().length > 0) return;

                    const width = doc.documentElement.clientWidth || window.innerWidth;
                    const x = e.clientX;

                    if (currentFlow === "paginated") {
                        if (x < width * 0.25) {
                            rendition.prev();
                        } else if (x > width * 0.75) {
                            rendition.next();
                        } else {
                            toggleHUD();
                        }
                    } else {
                        toggleHUD();
                    }
                });

                // Mouse move inside iframe resets HUD timer if near edges
                doc.addEventListener("mousemove", function(e) {
                    if (e.clientY < 70 || e.clientY > window.innerHeight - 60) {
                        showHUD();
                    }
                });
            });
        }

        // Render book engine
        function initRendition(targetCfi) {
            if (rendition) {
                try { rendition.destroy(); } catch(e) {}
                const viewerEl = document.getElementById("viewer");
                if (viewerEl) viewerEl.innerHTML = "";
            }

            document.body.classList.toggle("mode-scroll", currentFlow === "scrolled-doc");
            const spreadMode = window.innerWidth > 1050 ? "auto" : "none";

            rendition = book.renderTo("viewer", {
                width: "100%",
                height: "100%",
                spread: currentFlow === "scrolled-doc" ? "none" : spreadMode,
                flow: currentFlow === "scrolled-doc" ? "scrolled-doc" : "paginated"
            });

            applyRenditionStyles();
            bindRenditionHooks();
            rendition.on("relocated", onRelocated);

            if (targetCfi) {
                return rendition.display(targetCfi);
            } else {
                return rendition.display();
            }
        }

        // Fetch Book Binary
        fetch(EPUB_URL)
            .then(function(res) {
                if (!res.ok) throw new Error("Erro HTTP " + res.status + " ao carregar ficheiro EPUB.");
                return res.arrayBuffer();
            })
            .then(function(buffer) {
                if (!buffer || buffer.byteLength === 0) {
                    throw new Error("O ficheiro do livro est\u00e1 vazio ou inacess\u00edvel.");
                }
                book = ePub(buffer);

                // Load Navigation & TOC
                book.loaded.navigation.then(function(nav) {
                    tocData = nav.toc || [];
                    renderTOC(tocData);
                }).catch(function(e) { console.warn("TOC loading error:", e); });

                // Check saved progress from API or bookmark
                return fetch("/api/reader/progress/" + BOOK_ID)
                    .then(function(r) { return r.json(); })
                    .then(function(data) {
                        const savedTarget = (data && data.progress && data.progress.length > 5) ? data.progress : INITIAL_BOOKMARK;
                        return initRendition(savedTarget);
                    })
                    .catch(function() {
                        return initRendition(INITIAL_BOOKMARK || null);
                    });
            })
            .then(function() {
                hideLoading();
                resetHUDTimer();
                renderBookmarksList();

                // Generate locations asynchronously in background for scrubbing & metrics
                setTimeout(function() {
                    if (book && book.locations) {
                        book.locations.generate(1024).then(function() {
                            updateMetrics();
                        }).catch(function() {});
                    }
                }, 600);
            })
            .catch(function(err) {
                console.error("Reader init error:", err);
                showLoadingError(err.message || "Erro ao renderizar o livro.");
            });

        // Relocated Event Handler
        function onRelocated(location) {
            if (!location || !location.start) return;
            const cfi = location.start.cfi;
            updateMetrics(cfi);

            // Update Chapter labels
            if (location.start.href) {
                const found = findChapterByHref(location.start.href, tocData);
                if (found) {
                    const label = found.label ? found.label.trim() : "Capítulo";
                    document.getElementById("hud-chapter").textContent = label;
                    document.getElementById("reading-chapter-bottom").textContent = label;
                    highlightActiveTOC(found.href);
                }
            }

            // Save Progress Debounced
            saveProgressDebounced(cfi);
        }

        function updateMetrics(cfi) {
            let pct = 0;
            if (!cfi && rendition && rendition.location && rendition.location.start) {
                cfi = rendition.location.start.cfi;
            }

            if (book && book.locations && book.locations.total > 0 && cfi) {
                pct = Math.round(book.locations.percentageFromCfi(cfi) * 100) || 0;
                pct = Math.max(0, Math.min(100, pct));

                // Calculate Reading Time Remaining (~220 words per minute)
                const totalLocs = book.locations.total;
                const currentLoc = book.locations.locationFromCfi(cfi) || 0;
                const locsLeft = Math.max(0, totalLocs - currentLoc);
                const minutesLeft = Math.ceil((locsLeft * 140) / 220);

                let timeStr = "";
                if (minutesLeft >= 60) {
                    const h = Math.floor(minutesLeft / 60);
                    const m = minutesLeft % 60;
                    timeStr = "~" + h + "h " + m + "m";
                } else if (minutesLeft > 0) {
                    timeStr = "~" + minutesLeft + " min";
                } else {
                    timeStr = "Fim";
                }
                document.getElementById("reading-time-left").textContent = timeStr;
            }

            document.getElementById("progress-fill").style.width = pct + "%";
            document.getElementById("reading-percentage").textContent = pct + "%";
        }

        // Save Progress API
        function saveProgressDebounced(cfi) {
            updateSyncBadge(false, "A guardar");
            clearTimeout(syncTimeout);
            syncTimeout = setTimeout(function() {
                let pct = 0;
                if (book && book.locations && book.locations.total > 0) {
                    pct = Math.round(book.locations.percentageFromCfi(cfi) * 100);
                }
                fetch("/api/reader/progress/" + BOOK_ID, {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ progress: cfi, percentage: pct, device: "CWA Web Reader" })
                })
                .then(function(r) { return r.json(); })
                .then(function() { updateSyncBadge(true, "Guardado"); })
                .catch(function() { updateSyncBadge(false, "Offline"); });
            }, 600);
        }

        function updateSyncBadge(saved, text) {
            const badge = document.getElementById("sync-status");
            badge.className = "sync-badge" + (saved ? " saved" : "");
            badge.querySelector("span").textContent = text;
        }

        // Interactive Progress Scrubber Click & Drag
        const progressTrack = document.getElementById("progress-track");
        progressTrack.addEventListener("click", function(e) {
            if (!book || !book.locations || book.locations.total === 0 || !rendition) return;
            const rect = progressTrack.getBoundingClientRect();
            const ratio = Math.max(0, Math.min(1, (e.clientX - rect.left) / rect.width));
            const cfi = book.locations.cfiFromPercentage(ratio);
            if (cfi) rendition.display(cfi);
        });

        // Table of Contents
        function renderTOC(items) {
            const list = document.getElementById("toc-list");
            list.innerHTML = "";
            if (!items || items.length === 0) {
                list.innerHTML = '<li class="drawer-empty-state">Nenhum capítulo encontrado no índice.</li>';
                return;
            }

            function addItems(itemArray, depth) {
                itemArray.forEach(function(item) {
                    const li = document.createElement("li");
                    li.className = "toc-item";
                    li.dataset.href = item.href;
                    li.style.paddingLeft = (18 + depth * 14) + "px";
                    li.textContent = item.label ? item.label.trim() : "Sem título";
                    li.addEventListener("click", function() {
                        if (rendition) rendition.display(item.href);
                        closeDrawers();
                    });
                    list.appendChild(li);

                    if (item.subitems && item.subitems.length) {
                        addItems(item.subitems, depth + 1);
                    }
                });
            }

            addItems(items, 0);
        }

        function findChapterByHref(href, items) {
            if (!items) return null;
            for (let i = 0; i < items.length; i++) {
                const it = items[i];
                if (href.indexOf(it.href) !== -1 || it.href.indexOf(href) !== -1) return it;
                if (it.subitems && it.subitems.length) {
                    const sub = findChapterByHref(href, it.subitems);
                    if (sub) return sub;
                }
            }
            return null;
        }

        function highlightActiveTOC(href) {
            document.querySelectorAll(".toc-item").forEach(function(el) {
                if (el.dataset.href && (href.indexOf(el.dataset.href) !== -1 || el.dataset.href.indexOf(href) !== -1)) {
                    el.classList.add("active");
                } else {
                    el.classList.remove("active");
                }
            });
        }

        // TOC Filter
        document.getElementById("toc-filter-input").addEventListener("input", function(e) {
            const term = e.target.value.toLowerCase().trim();
            document.querySelectorAll(".toc-item").forEach(function(el) {
                el.style.display = el.textContent.toLowerCase().indexOf(term) !== -1 ? "" : "none";
            });
        });

        // In-Book Full-Text Search
        let isSearching = false;
        async function runBookSearch() {
            const input = document.getElementById("book-search-input");
            const query = input.value.trim();
            if (!book || query.length < 2 || isSearching) return;

            isSearching = true;
            const resultsList = document.getElementById("search-results-list");
            const statusBar = document.getElementById("search-status-bar");
            statusBar.style.display = "block";
            statusBar.textContent = "A pesquisar...";
            resultsList.innerHTML = '<li class="drawer-empty-state"><div class="spinner-ring" style="width:24px; height:24px; margin:0 auto 10px;"></div>A pesquisar no livro...</li>';

            try {
                const results = [];
                const spineItems = book.spine.spineItems;

                for (let i = 0; i < spineItems.length; i++) {
                    const item = spineItems[i];
                    await item.load(book.load.bind(book));
                    const found = item.find(query);
                    item.unload();

                    if (found && found.length) {
                        const ch = findChapterByHref(item.href, tocData);
                        const chName = ch ? ch.label.trim() : ("Capítulo " + (i + 1));
                        found.forEach(function(f) {
                            results.push({ cfi: f.cfi, excerpt: f.excerpt, chapter: chName });
                        });
                    }
                    if (results.length >= 60) break; // performance limit
                }

                statusBar.textContent = results.length + " ocorrências encontradas";
                resultsList.innerHTML = "";

                if (results.length === 0) {
                    resultsList.innerHTML = '<li class="drawer-empty-state">Nenhum resultado para "' + query + '".</li>';
                } else {
                    results.forEach(function(res) {
                        const li = document.createElement("li");
                        li.className = "search-item";
                        const escapedQuery = query.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
                        const highlighted = res.excerpt.replace(new RegExp("(" + escapedQuery + ")", "gi"), "<mark>$1</mark>");
                        li.innerHTML = '<div style="font-weight:600; font-size:12px; color:var(--accent-color);">' + res.chapter + '</div>' +
                            '<div class="search-item-snippet">' + highlighted + '</div>';
                        li.addEventListener("click", function() {
                            if (rendition) rendition.display(res.cfi);
                            closeDrawers();
                        });
                        resultsList.appendChild(li);
                    });
                }
            } catch (err) {
                resultsList.innerHTML = '<li class="drawer-empty-state">Erro na pesquisa do livro.</li>';
            } finally {
                isSearching = false;
            }
        }

        document.getElementById("btn-run-search").addEventListener("click", runBookSearch);
        document.getElementById("book-search-input").addEventListener("keydown", function(e) {
            if (e.key === "Enter") runBookSearch();
        });

        // Bookmarks Management
        function getBookmarksKey() { return "cwa_bookmarks_" + BOOK_ID; }

        function loadBookmarks() {
            try {
                return JSON.parse(localStorage.getItem(getBookmarksKey()) || "[]");
            } catch(e) { return []; }
        }

        function saveBookmarks(arr) {
            localStorage.setItem(getBookmarksKey(), JSON.stringify(arr));
            renderBookmarksList();
        }

        function addCurrentBookmark() {
            if (!rendition || !rendition.location || !rendition.location.start) return;
            const cfi = rendition.location.start.cfi;
            const chTitle = document.getElementById("hud-chapter").textContent || "Página";
            const pct = document.getElementById("reading-percentage").textContent || "0%";
            const bookmarks = loadBookmarks();

            // Avoid duplicate CFI bookmark
            if (!bookmarks.some(function(b) { return b.cfi === cfi; })) {
                bookmarks.unshift({
                    cfi: cfi,
                    title: chTitle,
                    percentage: pct,
                    date: new Date().toLocaleDateString("pt-PT", { day: "2-digit", month: "short", hour: "2-digit", minute: "2-digit" })
                });
                saveBookmarks(bookmarks);
                updateSyncBadge(true, "Marcador adicionado");
            }
        }

        function removeBookmark(index) {
            const bookmarks = loadBookmarks();
            bookmarks.splice(index, 1);
            saveBookmarks(bookmarks);
        }

        function renderBookmarksList() {
            const list = document.getElementById("bookmarks-list");
            const bookmarks = loadBookmarks();
            list.innerHTML = "";

            if (bookmarks.length === 0) {
                list.innerHTML = '<li class="drawer-empty-state">Nenhum marcador guardado neste livro.<br><small style="opacity:0.6;">Clica no ícone 📑 no topo para marcar páginas.</small></li>';
                return;
            }

            bookmarks.forEach(function(b, idx) {
                const li = document.createElement("li");
                li.className = "bookmark-item";
                li.innerHTML = '<div style="display:flex; justify-content:space-between; align-items:center;">' +
                    '<span style="font-weight:600; font-size:13px;">' + b.title + '</span>' +
                    '<button class="btn-icon" data-del-bm="' + idx + '" style="width:24px; height:24px; opacity:0.6;" title="Eliminar marcador">' +
                    '<svg viewBox="0 0 24 24"><path d="M18 6L6 18M6 6l12 12"/></svg></button>' +
                    '</div>' +
                    '<div class="bookmark-actions"><span>' + b.percentage + ' lido</span><span>' + b.date + '</span></div>';

                li.querySelector("[data-del-bm]").addEventListener("click", function(e) {
                    e.stopPropagation();
                    removeBookmark(idx);
                });

                li.addEventListener("click", function() {
                    if (rendition) rendition.display(b.cfi);
                    closeDrawers();
                });
                list.appendChild(li);
            });
        }

        document.getElementById("btn-bookmark").addEventListener("click", addCurrentBookmark);
        document.getElementById("btn-add-bookmark-drawer").addEventListener("click", addCurrentBookmark);

        // Drawer Navigation Tabs
        document.querySelectorAll(".drawer-tab").forEach(function(btn) {
            btn.addEventListener("click", function() {
                const tab = btn.dataset.tab;
                document.querySelectorAll(".drawer-tab").forEach(function(b) { b.classList.toggle("active", b === btn); });
                document.querySelectorAll(".drawer-content").forEach(function(c) {
                    c.classList.toggle("active", c.id === "tab-content-" + tab);
                });
            });
        });

        // Settings (Aa) Management
        function applyTheme(theme) {
            document.body.setAttribute("data-theme", theme);
            currentTheme = theme;
            localStorage.setItem("cwa_reader_theme", theme);
            document.querySelectorAll("[data-theme-btn]").forEach(function(b) {
                b.classList.toggle("active", b.dataset.themeBtn === theme);
            });
            applyRenditionStyles();
        }

        document.querySelectorAll("[data-theme-btn]").forEach(function(btn) {
            btn.addEventListener("click", function() { applyTheme(btn.dataset.themeBtn); });
        });

        function applyFont(font) {
            currentFont = font;
            localStorage.setItem("cwa_reader_font", font);
            document.querySelectorAll("[data-font-btn]").forEach(function(b) {
                b.classList.toggle("active", b.dataset.fontBtn === font);
            });
            applyRenditionStyles();
        }

        document.querySelectorAll("[data-font-btn]").forEach(function(btn) {
            btn.addEventListener("click", function() { applyFont(btn.dataset.fontBtn); });
        });

        // Font Size Controls
        document.getElementById("btn-font-inc").addEventListener("click", function() {
            if (currentFontSize < 220) {
                currentFontSize += 10;
                updateFontSize();
            }
        });
        document.getElementById("btn-font-dec").addEventListener("click", function() {
            if (currentFontSize > 70) {
                currentFontSize -= 10;
                updateFontSize();
            }
        });

        function updateFontSize() {
            document.getElementById("font-size-display").textContent = currentFontSize + "%";
            localStorage.setItem("cwa_reader_font_size", currentFontSize);
            applyRenditionStyles();
        }

        // Line Height Controls
        document.querySelectorAll("[data-lh-btn]").forEach(function(btn) {
            btn.addEventListener("click", function() {
                currentLineHeight = btn.dataset.lhBtn;
                localStorage.setItem("cwa_reader_lh", currentLineHeight);
                document.querySelectorAll("[data-lh-btn]").forEach(function(b) { b.classList.toggle("active", b === btn); });
                applyRenditionStyles();
            });
        });

        // Page Width Controls
        document.querySelectorAll("[data-width-btn]").forEach(function(btn) {
            btn.addEventListener("click", function() {
                currentPageWidth = btn.dataset.widthBtn;
                localStorage.setItem("cwa_reader_width", currentPageWidth);
                document.documentElement.style.setProperty("--page-width", currentPageWidth);
                document.querySelectorAll("[data-width-btn]").forEach(function(b) { b.classList.toggle("active", b === btn); });
                if (rendition) {
                    setTimeout(function() { rendition.resize(); }, 150);
                }
            });
        });

        // Text Alignment Controls
        document.querySelectorAll("[data-align-btn]").forEach(function(btn) {
            btn.addEventListener("click", function() {
                currentTextAlign = btn.dataset.alignBtn;
                localStorage.setItem("cwa_reader_align", currentTextAlign);
                document.querySelectorAll("[data-align-btn]").forEach(function(b) { b.classList.toggle("active", b === btn); });
                applyRenditionStyles();
            });
        });

        // Reading Flow Toggle (Paginated vs Scroll Contínuo)
        function toggleFlowMode() {
            currentFlow = currentFlow === "paginated" ? "scrolled-doc" : "paginated";
            localStorage.setItem("cwa_reader_flow", currentFlow);
            updateFlowIcon();

            if (rendition) {
                const curLoc = rendition.currentLocation();
                const cfi = (curLoc && curLoc.start) ? curLoc.start.cfi : null;
                initRendition(cfi);
            }
        }

        function updateFlowIcon() {
            const isPaginated = currentFlow === "paginated";
            document.getElementById("icon-flow-paginated").style.display = isPaginated ? "block" : "none";
            document.getElementById("icon-flow-scroll").style.display = isPaginated ? "none" : "block";
            const btn = document.getElementById("btn-toggle-flow");
            btn.title = isPaginated ? "Mudar para Scroll Cont\u00ednuo (M)" : "Mudar para Modo Paginado (M)";
        }

        document.getElementById("btn-toggle-flow").addEventListener("click", toggleFlowMode);

        // Chapter Next / Prev buttons
        function goToChapterOffset(offset) {
            if (!rendition || !rendition.location || !rendition.location.start || !tocData || tocData.length === 0) return;
            const currentHref = rendition.location.start.href;
            const flatToc = [];
            function flatten(arr) {
                arr.forEach(function(item) {
                    flatToc.push(item);
                    if (item.subitems && item.subitems.length) flatten(item.subitems);
                });
            }
            flatten(tocData);

            let currentIdx = -1;
            for (let i = 0; i < flatToc.length; i++) {
                if (currentHref.indexOf(flatToc[i].href) !== -1 || flatToc[i].href.indexOf(currentHref) !== -1) {
                    currentIdx = i;
                    break;
                }
            }

            const targetIdx = currentIdx + offset;
            if (targetIdx >= 0 && targetIdx < flatToc.length) {
                rendition.display(flatToc[targetIdx].href);
            }
        }

        document.getElementById("btn-prev-chapter").addEventListener("click", function() { goToChapterOffset(-1); });
        document.getElementById("btn-next-chapter").addEventListener("click", function() { goToChapterOffset(1); });

        // Edge chevrons
        document.getElementById("edge-prev").addEventListener("click", function() { if (rendition) rendition.prev(); });
        document.getElementById("edge-next").addEventListener("click", function() { if (rendition) rendition.next(); });

        // Drawer & Popover Toggles
        const drawer = document.getElementById("reader-drawer");
        const popover = document.getElementById("settings-popover");
        const backdrop = document.getElementById("backdrop");

        function openDrawer(tabName) {
            popover.classList.remove("open");
            drawer.classList.add("open");
            backdrop.classList.add("open");
            hudLocked = true;
            showHUD();

            if (tabName) {
                const tabBtn = document.querySelector('.drawer-tab[data-tab="' + tabName + '"]');
                if (tabBtn) tabBtn.click();
            }
        }

        function toggleSettings() {
            drawer.classList.remove("open");
            const isOpen = popover.classList.toggle("open");
            backdrop.classList.toggle("open", isOpen);
            hudLocked = isOpen;
            if (isOpen) showHUD();
            else resetHUDTimer();
        }

        function closeDrawers() {
            drawer.classList.remove("open");
            popover.classList.remove("open");
            backdrop.classList.remove("open");
            hudLocked = false;
            resetHUDTimer();
        }

        document.getElementById("btn-open-toc").addEventListener("click", function() { openDrawer("toc"); });
        document.getElementById("btn-close-drawer").addEventListener("click", closeDrawers);
        document.getElementById("btn-settings").addEventListener("click", toggleSettings);
        backdrop.addEventListener("click", closeDrawers);

        // Fullscreen Toggle
        function toggleFullscreen() {
            if (!document.fullscreenElement) {
                document.documentElement.requestFullscreen().catch(function() {});
            } else {
                document.exitFullscreen().catch(function() {});
            }
        }
        document.getElementById("btn-fullscreen").addEventListener("click", toggleFullscreen);

        // Global Keyboard Shortcuts
        document.addEventListener("keydown", function(e) {
            if (e.target.tagName === "INPUT" || e.target.tagName === "TEXTAREA") {
                if (e.key === "Escape") closeDrawers();
                return;
            }
            if (e.key === "ArrowLeft" || e.key === "j" || e.key === "PageUp") {
                if (rendition) rendition.prev();
            } else if (e.key === "ArrowRight" || e.key === "k" || e.key === " " || e.key === "PageDown") {
                if (rendition) rendition.next();
            } else if (e.key === "t" || e.key === "T") {
                openDrawer("toc");
            } else if (e.key === "s" || e.key === "S" || e.key === "a" || e.key === "A") {
                toggleSettings();
            } else if (e.key === "b" || e.key === "B") {
                addCurrentBookmark();
            } else if (e.key === "m" || e.key === "M") {
                toggleFlowMode();
            } else if (e.key === "f" || e.key === "F") {
                toggleFullscreen();
            } else if (e.key === "/" || (e.ctrlKey && e.key === "f")) {
                e.preventDefault();
                openDrawer("search");
                setTimeout(function() { document.getElementById("book-search-input").focus(); }, 150);
            } else if (e.key === "Escape") {
                closeDrawers();
            }
        });

        // Window resize adjusts spread & rendition layout
        let resizeTimer = null;
        window.addEventListener("resize", function() {
            clearTimeout(resizeTimer);
            resizeTimer = setTimeout(function() {
                if (rendition) {
                    if (currentFlow === "paginated") {
                        const spread = window.innerWidth > 1050 ? "auto" : "none";
                        rendition.spread(spread);
                    }
                    rendition.resize();
                }
            }, 200);
        });
    </script>
</body>
</html>`


