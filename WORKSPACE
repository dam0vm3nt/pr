load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "16e9fca53ed6bd4ff4ad76facc9b7b651a89db1689a2877d6fd7b82aa824e366",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.34.0/rules_go-v0.34.0.zip",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.34.0/rules_go-v0.34.0.zip",
    ],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "5982e5463f171da99e3bdaeff8c0f48283a7a5f396ec5282910b9e8a49c0dd7e",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.25.0/bazel-gazelle-v0.25.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.25.0/bazel-gazelle-v0.25.0.tar.gz",
    ],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")

go_repository(
    name = "co_honnef_go_tools",
    importpath = "honnef.co/go/tools",
    sum = "h1:UoveltGrhghAA7ePc+e+QYDHXrBps2PqFZiHkGR/xK8=",
    version = "v0.0.1-2020.1.4",
)

go_repository(
    name = "com_github_acomagu_bufpipe",
    importpath = "github.com/acomagu/bufpipe",
    sum = "h1:fxAGrHZTgQ9w5QqVItgzwj235/uYZYgbXitB+dLupOk=",
    version = "v1.0.3",
)

go_repository(
    name = "com_github_alecaivazis_survey_v2",
    importpath = "github.com/AlecAivazis/survey/v2",
    sum = "h1:pchTU9rsLUSvWEl2Aq9Pv3k0IE2fkqtGxazskAMd9Ng=",
    version = "v2.3.4",
)

go_repository(
    name = "com_github_alecthomas_chroma",
    importpath = "github.com/alecthomas/chroma",
    sum = "h1:7XDcGkCQopCNKjZHfYrNLraA+M7e0fMiJ/Mfikbfjek=",
    version = "v0.10.0",
)

go_repository(
    name = "com_github_anmitsu_go_shlex",
    importpath = "github.com/anmitsu/go-shlex",
    sum = "h1:kFOfPq6dUM1hTo4JG6LR5AXSUEsOjtdm0kw0FtQtMJA=",
    version = "v0.0.0-20161002113705-648efa622239",
)

go_repository(
    name = "com_github_antihax_optional",
    importpath = "github.com/antihax/optional",
    sum = "h1:xK2lYat7ZLaVVcIuj82J8kIro4V6kDe0AUDFboUCwcg=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_armon_go_socks5",
    importpath = "github.com/armon/go-socks5",
    sum = "h1:0CwZNZbxp69SHPdPJAN/hZIm0C4OItdklCFmMRWYpio=",
    version = "v0.0.0-20160902184237-e75332964ef5",
)

go_repository(
    name = "com_github_atomicgo_cursor",
    importpath = "github.com/atomicgo/cursor",
    sum = "h1:xdogsqa6YYlLfM+GyClC/Lchf7aiMerFiZQn7soTOoU=",
    version = "v0.0.1",
)

go_repository(
    name = "com_github_atotto_clipboard",
    importpath = "github.com/atotto/clipboard",
    sum = "h1:EH0zSVneZPSuFR11BlR9YppQTVDbh5+16AmcJi4g1z4=",
    version = "v0.1.4",
)

go_repository(
    name = "com_github_aymerick_douceur",
    importpath = "github.com/aymerick/douceur",
    sum = "h1:Mv+mAeH1Q+n9Fr+oyamOlAkUNPWPlA8PPGR0QAaYuPk=",
    version = "v0.2.0",
)

go_repository(
    name = "com_github_bluekeyes_go_gitdiff",
    importpath = "github.com/bluekeyes/go-gitdiff",
    sum = "h1:VPpQZanEVwFid0IoAIyiJLD+SMWM/AXZXYFc+3E9CS4=",
    version = "v0.6.1",
)

go_repository(
    name = "com_github_bradleyfalzon_ghinstallation_v2",
    importpath = "github.com/bradleyfalzon/ghinstallation/v2",
    sum = "h1:tXKVfhE7FcSkhkv0UwkLvPDeZ4kz6OXd0PKPlFqf81M=",
    version = "v2.0.4",
)

go_repository(
    name = "com_github_briandowns_spinner",
    importpath = "github.com/briandowns/spinner",
    sum = "h1:yhQmQtM1zsqFsouh09Bk/jCjd50pC3EOGsh28gLVvwY=",
    version = "v1.18.1",
)

go_repository(
    name = "com_github_burntsushi_toml",
    importpath = "github.com/BurntSushi/toml",
    sum = "h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=",
    version = "v0.3.1",
)

go_repository(
    name = "com_github_burntsushi_xgb",
    importpath = "github.com/BurntSushi/xgb",
    sum = "h1:1BDTz0u9nC3//pOCMdNH+CiXJVYJh5UQNCOBG7jbELc=",
    version = "v0.0.0-20160522181843-27f122750802",
)

go_repository(
    name = "com_github_census_instrumentation_opencensus_proto",
    importpath = "github.com/census-instrumentation/opencensus-proto",
    sum = "h1:glEXhBS5PSLLv4IXzLA5yPRVX4bilULVyxxbrfOtDAk=",
    version = "v0.2.1",
)

go_repository(
    name = "com_github_charmbracelet_bubbles",
    importpath = "github.com/charmbracelet/bubbles",
    sum = "h1:fBLyY0PvJnd56Vlu5L84JJH6f4axhgIJ9P3NET78f0Q=",
    version = "v0.11.0",
)

go_repository(
    name = "com_github_charmbracelet_bubbletea",
    importpath = "github.com/charmbracelet/bubbletea",
    sum = "h1:f3y+kanzgev5PA916qxmDybSHU3N804uOnKnhRPXTcI=",
    version = "v0.21.0",
)

go_repository(
    name = "com_github_charmbracelet_glamour",
    importpath = "github.com/charmbracelet/glamour",
    sum = "h1:wu15ykPdB7X6chxugG/NNfDUbyyrCLV9XBalj5wdu3g=",
    version = "v0.5.0",
)

go_repository(
    name = "com_github_charmbracelet_harmonica",
    importpath = "github.com/charmbracelet/harmonica",
    sum = "h1:8NxJWRWg/bzKqqEaaeFNipOu77YR5t8aSwG4pgaUBiQ=",
    version = "v0.2.0",
)

go_repository(
    name = "com_github_charmbracelet_lipgloss",
    importpath = "github.com/charmbracelet/lipgloss",
    sum = "h1:lulQHuVeodSgDez+3rGiuxlPVXSnhth442DATR2/8t8=",
    version = "v0.5.0",
)

go_repository(
    name = "com_github_chzyer_logex",
    importpath = "github.com/chzyer/logex",
    sum = "h1:Swpa1K6QvQznwJRcfTfQJmTE72DqScAa40E+fbHEXEE=",
    version = "v1.1.10",
)

go_repository(
    name = "com_github_chzyer_readline",
    importpath = "github.com/chzyer/readline",
    sum = "h1:fY5BOSpyZCqRo5OhCuC+XN+r/bBCmeuuJtjz+bCNIf8=",
    version = "v0.0.0-20180603132655-2972be24d48e",
)

go_repository(
    name = "com_github_chzyer_test",
    importpath = "github.com/chzyer/test",
    sum = "h1:q763qf9huN11kDQavWsoZXJNW3xEE4JJyHa5Q25/sd8=",
    version = "v0.0.0-20180213035817-a1ea475d72b1",
)

go_repository(
    name = "com_github_cli_browser",
    importpath = "github.com/cli/browser",
    sum = "h1:xOZBfkfY9L9vMBgqb1YwRirGu6QFaQ5dP/vXt5ENSOY=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_cli_cli_v2",
    importpath = "github.com/cli/cli/v2",
    sum = "h1:IjzBU+VLKmR3ZrFpZcLKjntf3d+oLY2OnIDutrC1S4E=",
    version = "v2.8.0",
)

go_repository(
    name = "com_github_cli_oauth",
    importpath = "github.com/cli/oauth",
    sum = "h1:nxBC0Df4tUzMkqffAB+uZvisOwT3/N9FpkfdTDtafxc=",
    version = "v0.9.0",
)

go_repository(
    name = "com_github_cli_safeexec",
    importpath = "github.com/cli/safeexec",
    sum = "h1:0VngyaIyqACHdcMNWfo6+KdUYnqEr2Sg+bSP1pdF+dI=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_cli_shurcool_graphql",
    importpath = "github.com/cli/shurcooL-graphql",
    sum = "h1:/9J3t9O6p1B8zdBBtQighq5g7DQRItBwuwGh3SocsKM=",
    version = "v0.0.1",
)

go_repository(
    name = "com_github_client9_misspell",
    importpath = "github.com/client9/misspell",
    sum = "h1:ta993UF76GwbvJcIo3Y68y/M3WxlpEHPWIGDkJYwzJI=",
    version = "v0.3.4",
)

go_repository(
    name = "com_github_cncf_udpa_go",
    importpath = "github.com/cncf/udpa/go",
    sum = "h1:WBZRG4aNOuI15bLRrCgN8fCq8E5Xuty6jGbmSNEvSsU=",
    version = "v0.0.0-20191209042840-269d4d468f6f",
)

go_repository(
    name = "com_github_containerd_console",
    importpath = "github.com/containerd/console",
    sum = "h1:lIr7SlA5PxZyMV30bDW0MGbiOPXwc63yRuCP0ARubLw=",
    version = "v1.0.3",
)

go_repository(
    name = "com_github_cpuguy83_go_md2man_v2",
    importpath = "github.com/cpuguy83/go-md2man/v2",
    sum = "h1:r/myEWzV9lfsM1tFLgDyu0atFtJ1fXn261LKYj/3DxU=",
    version = "v2.0.1",
)

go_repository(
    name = "com_github_creack_pty",
    importpath = "github.com/creack/pty",
    sum = "h1:n56/Zwd5o6whRC5PMGretI4IdRLlmBXYNjScPaBgsbY=",
    version = "v1.1.18",
)

go_repository(
    name = "com_github_davecgh_go_spew",
    importpath = "github.com/davecgh/go-spew",
    sum = "h1:vj9j/u1bqnvCEfJOwUhtlOARqs3+rkHYY13jYWTU97c=",
    version = "v1.1.1",
)

go_repository(
    name = "com_github_dlclark_regexp2",
    importpath = "github.com/dlclark/regexp2",
    sum = "h1:F1rxgk7p4uKjwIQxBs9oAXe5CqrXlCduYEJvrF4u93E=",
    version = "v1.4.0",
)

go_repository(
    name = "com_github_emirpasic_gods",
    importpath = "github.com/emirpasic/gods",
    sum = "h1:QAUIPSaCu4G+POclxeqb3F+WPpdKqFGlw36+yOzGlrg=",
    version = "v1.12.0",
)

go_repository(
    name = "com_github_envoyproxy_go_control_plane",
    importpath = "github.com/envoyproxy/go-control-plane",
    sum = "h1:rEvIZUSZ3fx39WIi3JkQqQBitGwpELBIYWeBVh6wn+E=",
    version = "v0.9.4",
)

go_repository(
    name = "com_github_envoyproxy_protoc_gen_validate",
    importpath = "github.com/envoyproxy/protoc-gen-validate",
    sum = "h1:EQciDnbrYxy13PgWoY8AqoxGiPrpgBZ1R8UNe3ddc+A=",
    version = "v0.1.0",
)

go_repository(
    name = "com_github_erikgeiser_promptkit",
    importpath = "github.com/erikgeiser/promptkit",
    sum = "h1:Cxw/MBLZ+dxF7iOHMi/Z8dSi1vZeobPJsx8phAnsgW4=",
    version = "v0.6.0",
)

go_repository(
    name = "com_github_fatih_color",
    importpath = "github.com/fatih/color",
    sum = "h1:DkWD4oS2D8LGGgTQ6IvwJJXSL5Vp2ffcQg58nFV38Ys=",
    version = "v1.7.0",
)

go_repository(
    name = "com_github_flynn_go_shlex",
    importpath = "github.com/flynn/go-shlex",
    sum = "h1:BHsljHzVlRcyQhjrss6TZTdY2VfCqZPbv5k3iBFa2ZQ=",
    version = "v0.0.0-20150515145356-3f9db97f8568",
)

go_repository(
    name = "com_github_gabriel_vasile_mimetype",
    importpath = "github.com/gabriel-vasile/mimetype",
    sum = "h1:Cn9dkdYsMIu56tGho+fqzh7XmvY2YyGU0FnbhiOsEro=",
    version = "v1.4.0",
)

go_repository(
    name = "com_github_gliderlabs_ssh",
    importpath = "github.com/gliderlabs/ssh",
    sum = "h1:6zsha5zo/TWhRhwqCD3+EarCAgZ2yN28ipRnGPnwkI0=",
    version = "v0.2.2",
)

go_repository(
    name = "com_github_go_git_gcfg",
    importpath = "github.com/go-git/gcfg",
    sum = "h1:Q5ViNfGF8zFgyJWPqYwA7qGFoMTEiBmdlkcfRmpIMa4=",
    version = "v1.5.0",
)

go_repository(
    name = "com_github_go_git_go_billy_v5",
    importpath = "github.com/go-git/go-billy/v5",
    sum = "h1:CPiOUAzKtMRvolEKw+bG1PLRpT7D3LIs3/3ey4Aiu34=",
    version = "v5.3.1",
)

go_repository(
    name = "com_github_go_git_go_git_fixtures_v4",
    importpath = "github.com/go-git/go-git-fixtures/v4",
    sum = "h1:n9gGL1Ct/yIw+nfsfr8s4+sbhT+Ncu2SubfXjIWgci8=",
    version = "v4.2.1",
)

go_repository(
    name = "com_github_go_git_go_git_v5",
    importpath = "github.com/go-git/go-git/v5",
    sum = "h1:BXyZu9t0VkbiHtqrsvdq39UDhGJTl1h55VW6CSC4aY4=",
    version = "v5.4.2",
)

go_repository(
    name = "com_github_go_gl_glfw",
    importpath = "github.com/go-gl/glfw",
    sum = "h1:QbL/5oDUmRBzO9/Z7Seo6zf912W/a6Sr4Eu0G/3Jho0=",
    version = "v0.0.0-20190409004039-e6da0acd62b1",
)

go_repository(
    name = "com_github_go_gl_glfw_v3_3_glfw",
    importpath = "github.com/go-gl/glfw/v3.3/glfw",
    sum = "h1:WtGNWLvXpe6ZudgnXrq0barxBImvnnJoMEhXAzcbM0I=",
    version = "v0.0.0-20200222043503-6f7a984d4dc4",
)

go_repository(
    name = "com_github_golang_glog",
    importpath = "github.com/golang/glog",
    sum = "h1:VKtxabqXZkF25pY9ekfRL6a582T4P37/31XEstQ5p58=",
    version = "v0.0.0-20160126235308-23def4e6c14b",
)

go_repository(
    name = "com_github_golang_groupcache",
    importpath = "github.com/golang/groupcache",
    sum = "h1:1r7pUrabqp18hOBcwBwiTsbnFeTZHV9eER/QT5JVZxY=",
    version = "v0.0.0-20200121045136-8c9f03a8e57e",
)

go_repository(
    name = "com_github_golang_jwt_jwt_v4",
    importpath = "github.com/golang-jwt/jwt/v4",
    sum = "h1:RAqyYixv1p7uEnocuy8P1nru5wprCh/MH2BIlW5z5/o=",
    version = "v4.0.0",
)

go_repository(
    name = "com_github_golang_mock",
    importpath = "github.com/golang/mock",
    sum = "h1:l75CXGRSwbaYNpl/Z2X1XIIAMSCquvXgpVZDhwEIJsc=",
    version = "v1.4.4",
)

go_repository(
    name = "com_github_golang_protobuf",
    importpath = "github.com/golang/protobuf",
    sum = "h1:ROPKBNFfQgOUMifHyP+KYbvpjbdoFNs+aK7DXlji0Tw=",
    version = "v1.5.2",
)

go_repository(
    name = "com_github_google_btree",
    importpath = "github.com/google/btree",
    sum = "h1:0udJVsspx3VBr5FwtLhQQtuAsVc79tTq0ocGIPAU6qo=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_google_go_cmp",
    importpath = "github.com/google/go-cmp",
    sum = "h1:81/ik6ipDQS2aGcBfIN5dHDB36BwrStyeAQquSYCV4o=",
    version = "v0.5.7",
)

go_repository(
    name = "com_github_google_go_github_v41",
    importpath = "github.com/google/go-github/v41",
    sum = "h1:HseJrM2JFf2vfiZJ8anY2hqBjdfY1Vlj/K27ueww4gg=",
    version = "v41.0.0",
)

go_repository(
    name = "com_github_google_go_github_v43",
    importpath = "github.com/google/go-github/v43",
    sum = "h1:y+GL7LIsAIF2NZlJ46ZoC/D1W1ivZasT0lnWHMYPZ+U=",
    version = "v43.0.0",
)

go_repository(
    name = "com_github_google_go_querystring",
    importpath = "github.com/google/go-querystring",
    sum = "h1:AnCroh3fv4ZBgVIf1Iwtovgjaw/GiKJo8M8yD/fhyJ8=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_google_goterm",
    importpath = "github.com/google/goterm",
    sum = "h1:5CjVwnuUcp5adK4gmY6i72gpVFVnZDP2h5TmPScB6u4=",
    version = "v0.0.0-20190703233501-fc88cf888a3f",
)

go_repository(
    name = "com_github_google_martian",
    importpath = "github.com/google/martian",
    sum = "h1:/CP5g8u/VJHijgedC/Legn3BAbAaWPgecwXBIDzw5no=",
    version = "v2.1.0+incompatible",
)

go_repository(
    name = "com_github_google_martian_v3",
    importpath = "github.com/google/martian/v3",
    sum = "h1:pMen7vLs8nvgEYhywH3KDWJIJTeEr2ULsVWHWYHQyBs=",
    version = "v3.0.0",
)

go_repository(
    name = "com_github_google_pprof",
    importpath = "github.com/google/pprof",
    sum = "h1:Ak8CrdlwwXwAZxzS66vgPt4U8yUZX7JwLvVR58FN5jM=",
    version = "v0.0.0-20200708004538-1a94d8640e99",
)

go_repository(
    name = "com_github_google_renameio",
    importpath = "github.com/google/renameio",
    sum = "h1:GOZbcHa3HfsPKPlmyPyN2KEohoMXOhdMbHrvbpl2QaA=",
    version = "v0.1.0",
)

go_repository(
    name = "com_github_google_shlex",
    importpath = "github.com/google/shlex",
    sum = "h1:El6M4kTTCOh6aBiKaUGG7oYTSPP8MxqL4YI3kZKwcP4=",
    version = "v0.0.0-20191202100458-e7afc7fbc510",
)

go_repository(
    name = "com_github_googleapis_gax_go_v2",
    importpath = "github.com/googleapis/gax-go/v2",
    sum = "h1:sjZBwGj9Jlw33ImPtvFviGYvseOtDM7hkSKB7+Tv3SM=",
    version = "v2.0.5",
)

go_repository(
    name = "com_github_gookit_color",
    importpath = "github.com/gookit/color",
    sum = "h1:1Opow3+BWDwqor78DcJkJCIwnkviFi+rrOANki9BUFw=",
    version = "v1.5.0",
)

go_repository(
    name = "com_github_gorilla_css",
    importpath = "github.com/gorilla/css",
    sum = "h1:BQqNyPTi50JCFMTw/b67hByjMVXZRwGha6wxVGkeihY=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_gorilla_websocket",
    importpath = "github.com/gorilla/websocket",
    sum = "h1:+/TMaTYc4QFitKJxsQ7Yye35DkWvkdLcvGKqM+x0Ufc=",
    version = "v1.4.2",
)

go_repository(
    name = "com_github_hashicorp_errwrap",
    importpath = "github.com/hashicorp/errwrap",
    sum = "h1:hLrqtEDnRye3+sgx6z4qVLNuviH3MR5aQ0ykNJa/UYA=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_hashicorp_go_multierror",
    importpath = "github.com/hashicorp/go-multierror",
    sum = "h1:H5DkEtf6CXdFp0N0Em5UCwQpXMWke8IA0+lD48awMYo=",
    version = "v1.1.1",
)

go_repository(
    name = "com_github_hashicorp_go_version",
    importpath = "github.com/hashicorp/go-version",
    sum = "h1:McDWVJIU/y+u1BRV06dPaLfLCaT7fUTJLp5r04x7iNw=",
    version = "v1.3.0",
)

go_repository(
    name = "com_github_hashicorp_golang_lru",
    importpath = "github.com/hashicorp/golang-lru",
    sum = "h1:0hERBMJE1eitiLkihrMvRVBYAkpHzc/J3QdDN+dAcgU=",
    version = "v0.5.1",
)

go_repository(
    name = "com_github_henvic_httpretty",
    importpath = "github.com/henvic/httpretty",
    sum = "h1:JdzGzKZBajBfnvlMALXXMVQWxWMF/ofTy8C3/OSUTxs=",
    version = "v0.0.6",
)

go_repository(
    name = "com_github_hinshun_vt10x",
    importpath = "github.com/hinshun/vt10x",
    sum = "h1:qv2VnGeEQHchGaZ/u7lxST/RaJw+cv273q79D81Xbog=",
    version = "v0.0.0-20220119200601-820417d04eec",
)

go_repository(
    name = "com_github_ianlancetaylor_demangle",
    importpath = "github.com/ianlancetaylor/demangle",
    sum = "h1:UDMh68UUwekSh5iP2OMhRRZJiiBccgV7axzUG8vi56c=",
    version = "v0.0.0-20181102032728-5e5cf60278f6",
)

go_repository(
    name = "com_github_imdario_mergo",
    importpath = "github.com/imdario/mergo",
    sum = "h1:b6R2BslTbIEToALKP7LxUvijTsNI9TAe80pLWN2g/HU=",
    version = "v0.3.12",
)

go_repository(
    name = "com_github_inconshreveable_mousetrap",
    importpath = "github.com/inconshreveable/mousetrap",
    sum = "h1:Z8tu5sraLXCXIcARxBp/8cbvlwVa7Z1NHg9XEKhtSvM=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_itchyny_gojq",
    importpath = "github.com/itchyny/gojq",
    sum = "h1:hYPTpeWfrJ1OT+2j6cvBScbhl0TkdwGM4bc66onUSOQ=",
    version = "v0.12.7",
)

go_repository(
    name = "com_github_itchyny_timefmt_go",
    importpath = "github.com/itchyny/timefmt-go",
    sum = "h1:7M3LGVDsqcd0VZH2U+x393obrzZisp7C0uEe921iRkU=",
    version = "v0.1.3",
)

go_repository(
    name = "com_github_jbenet_go_context",
    importpath = "github.com/jbenet/go-context",
    sum = "h1:BQSFePA1RWJOlocH6Fxy8MmwDt+yVQYULKfN0RoTN8A=",
    version = "v0.0.0-20150711004518-d14ea06fba99",
)

go_repository(
    name = "com_github_jessevdk_go_flags",
    importpath = "github.com/jessevdk/go-flags",
    sum = "h1:1jKYvbxEjfUl0fmqTCOfonvskHHXMjBySTLW4y9LFvc=",
    version = "v1.5.0",
)

go_repository(
    name = "com_github_joho_godotenv",
    importpath = "github.com/joho/godotenv",
    sum = "h1:3l4+N6zfMWnkbPEXKng2o2/MR5mSwTrBih4ZEkkz1lg=",
    version = "v1.4.0",
)

go_repository(
    name = "com_github_jstemmer_go_junit_report",
    importpath = "github.com/jstemmer/go-junit-report",
    sum = "h1:6QPYqodiu3GuPL+7mfx+NwDdp2eTkp9IfEUpgAwUN0o=",
    version = "v0.9.1",
)

go_repository(
    name = "com_github_kballard_go_shellquote",
    importpath = "github.com/kballard/go-shellquote",
    sum = "h1:Z9n2FFNUXsshfwJMBgNA0RU6/i7WVaAegv3PtuIHPMs=",
    version = "v0.0.0-20180428030007-95032a82bc51",
)

go_repository(
    name = "com_github_kevinburke_ssh_config",
    importpath = "github.com/kevinburke/ssh_config",
    sum = "h1:DowS9hvgyYSX4TO5NpyC606/Z4SxnNYbT+WX27or6Ck=",
    version = "v0.0.0-20201106050909-4977a11b4351",
)

go_repository(
    name = "com_github_kisielk_gotool",
    importpath = "github.com/kisielk/gotool",
    sum = "h1:AV2c/EiW3KqPNT9ZKl07ehoAGi4C5/01Cfbblndcapg=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_klauspost_cpuid_v2",
    importpath = "github.com/klauspost/cpuid/v2",
    sum = "h1:p9dKCg8i4gmOxtv35DvrYoWqYzQrvEVdjQ762Y0OqZE=",
    version = "v2.0.12",
)

go_repository(
    name = "com_github_konsorten_go_windows_terminal_sequences",
    importpath = "github.com/konsorten/go-windows-terminal-sequences",
    sum = "h1:mweAR1A6xJ3oS2pRaGiHgQ4OO8tzTaLawm8vnODuwDk=",
    version = "v1.0.1",
)

go_repository(
    name = "com_github_kr_pretty",
    importpath = "github.com/kr/pretty",
    sum = "h1:Fmg33tUaq4/8ym9TJN1x7sLJnHVwhP33CNkpYV/7rwI=",
    version = "v0.2.1",
)

go_repository(
    name = "com_github_kr_pty",
    importpath = "github.com/kr/pty",
    sum = "h1:VkoXIwSboBpnk99O/KFauAEILuNHv5DVFKZMBN/gUgw=",
    version = "v1.1.1",
)

go_repository(
    name = "com_github_kr_text",
    importpath = "github.com/kr/text",
    sum = "h1:5Nx0Ya0ZqY2ygV366QzturHI13Jq95ApcVaJBhpS+AY=",
    version = "v0.2.0",
)

go_repository(
    name = "com_github_kylelemons_godebug",
    importpath = "github.com/kylelemons/godebug",
    sum = "h1:RPNrshWIDI6G2gRW9EHilWtl7Z6Sb1BR0xunSBf0SNc=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_lucasb_eyer_go_colorful",
    importpath = "github.com/lucasb-eyer/go-colorful",
    sum = "h1:1nnpGOrhyZZuNyfu1QjKiUICQ74+3FNCN69Aj6K7nkY=",
    version = "v1.2.0",
)

go_repository(
    name = "com_github_makenowjust_heredoc",
    importpath = "github.com/MakeNowJust/heredoc",
    sum = "h1:cXCdzVdstXyiTqTvfqk9SDHpKNjxuom+DOlyEeQ4pzQ=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_marvinjwendt_testza",
    importpath = "github.com/MarvinJWendt/testza",
    sum = "h1:iLAixjb8IXSgiHOk9aOITXFP33k+D4qj9ZX5PXNKO5A=",
    version = "v0.3.0",
)

go_repository(
    name = "com_github_matryer_is",
    importpath = "github.com/matryer/is",
    sum = "h1:92UTHpy8CDwaJ08GqLDzhhuixiBUUD1p3AU6PHddz4A=",
    version = "v1.2.0",
)

go_repository(
    name = "com_github_mattn_go_colorable",
    importpath = "github.com/mattn/go-colorable",
    sum = "h1:jF+Du6AlPIjs2BiUiQlKOX0rt3SujHxPnksPKZbaA40=",
    version = "v0.1.12",
)

go_repository(
    name = "com_github_mattn_go_isatty",
    importpath = "github.com/mattn/go-isatty",
    sum = "h1:yVuAays6BHfxijgZPzw+3Zlu5yQgKGP2/hcQbHb7S9Y=",
    version = "v0.0.14",
)

go_repository(
    name = "com_github_mattn_go_runewidth",
    importpath = "github.com/mattn/go-runewidth",
    sum = "h1:lTGmDsbAYt5DmK6OnoV7EuIF1wEIFAcxld6ypU4OSgU=",
    version = "v0.0.13",
)

go_repository(
    name = "com_github_mgutz_ansi",
    importpath = "github.com/mgutz/ansi",
    sum = "h1:5PJl274Y63IEHC+7izoQE9x6ikvDFZS2mDVS3drnohI=",
    version = "v0.0.0-20200706080929-d51e80ef957d",
)

go_repository(
    name = "com_github_microcosm_cc_bluemonday",
    importpath = "github.com/microcosm-cc/bluemonday",
    sum = "h1:Z1a//hgsQ4yjC+8zEkV8IWySkXnsxmdSY642CTFQb5Y=",
    version = "v1.0.17",
)

go_repository(
    name = "com_github_microsoft_go_winio",
    importpath = "github.com/Microsoft/go-winio",
    sum = "h1:FtSW/jqD+l4ba5iPBj9CODVtgfYAD8w2wS923g/cFDk=",
    version = "v0.4.16",
)

go_repository(
    name = "com_github_mitchellh_go_homedir",
    importpath = "github.com/mitchellh/go-homedir",
    sum = "h1:lukF9ziXFxDFPkA1vsr5zpc1XuPDn/wFntq5mG+4E0Y=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_muesli_ansi",
    importpath = "github.com/muesli/ansi",
    sum = "h1:1XF24mVaiu7u+CFywTdcDo2ie1pzzhwjt6RHqzpMU34=",
    version = "v0.0.0-20211018074035-2e021307bc4b",
)

go_repository(
    name = "com_github_muesli_reflow",
    importpath = "github.com/muesli/reflow",
    sum = "h1:IFsN6K9NfGtjeggFP+68I4chLZV2yIKsXJFNZ+eWh6s=",
    version = "v0.3.0",
)

go_repository(
    name = "com_github_muesli_termenv",
    importpath = "github.com/muesli/termenv",
    sum = "h1:QANkGiGr39l1EESqrE0gZw0/AJNYzIvoGLhIoVYtluI=",
    version = "v0.11.1-0.20220212125758-44cd13922739",
)

go_repository(
    name = "com_github_muhammadmuzzammil1998_jsonc",
    importpath = "github.com/muhammadmuzzammil1998/jsonc",
    sum = "h1:0FrBxrkJ0hVembTb/e4EU5Ml6vLcOusAqymmYISg5Uo=",
    version = "v0.0.0-20201229145248-615b0916ca38",
)

go_repository(
    name = "com_github_netflix_go_expect",
    importpath = "github.com/Netflix/go-expect",
    sum = "h1:+vx7roKuyA63nhn5WAunQHLTznkw5W8b1Xc0dNjp83s=",
    version = "v0.0.0-20220104043353-73e0943537d2",
)

go_repository(
    name = "com_github_niemeyer_pretty",
    importpath = "github.com/niemeyer/pretty",
    sum = "h1:fD57ERR4JtEqsWbfPhv4DMiApHyliiK5xCTNVSPiaAs=",
    version = "v0.0.0-20200227124842-a10e7caefd8e",
)

go_repository(
    name = "com_github_olekukonko_tablewriter",
    importpath = "github.com/olekukonko/tablewriter",
    sum = "h1:P2Ga83D34wi1o9J6Wh1mRuqd4mF/x/lgBS7N7AbDhec=",
    version = "v0.0.5",
)

go_repository(
    name = "com_github_opentracing_opentracing_go",
    importpath = "github.com/opentracing/opentracing-go",
    sum = "h1:pWlfV3Bxv7k65HYwkikxat0+s3pV4bsqf19k25Ur8rU=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_pkg_errors",
    importpath = "github.com/pkg/errors",
    sum = "h1:FEBLx1zS214owpjy7qsBeixbURkuhQAwrK5UwLGTwt4=",
    version = "v0.9.1",
)

go_repository(
    name = "com_github_pmezard_go_difflib",
    importpath = "github.com/pmezard/go-difflib",
    sum = "h1:4DBwDE0NGyQoBHbLQYPwSUPoCMWR5BEzIk/f1lZbAQM=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_prometheus_client_model",
    importpath = "github.com/prometheus/client_model",
    sum = "h1:gQz4mCbXsO+nc9n1hCxHcGA3Zx3Eo+UHZoInFGUIXNM=",
    version = "v0.0.0-20190812154241-14fe0d1b01d4",
)

go_repository(
    name = "com_github_protonmail_go_crypto",
    importpath = "github.com/ProtonMail/go-crypto",
    sum = "h1:YoJbenK9C67SkzkDfmQuVln04ygHj3vjZfd9FL+GmQQ=",
    version = "v0.0.0-20210428141323-04723f9f07d7",
)

go_repository(
    name = "com_github_pterm_pterm",
    importpath = "github.com/pterm/pterm",
    sum = "h1:LvQE43RYegVH+y5sCDcqjlbsRu0DlAecEn9FDfs9ePs=",
    version = "v0.12.40",
)

go_repository(
    name = "com_github_rivo_uniseg",
    importpath = "github.com/rivo/uniseg",
    sum = "h1:S1pD9weZBuJdFmowNwbpi7BJ8TNftyUImj/0WQi72jY=",
    version = "v0.2.0",
)

go_repository(
    name = "com_github_rogpeppe_go_internal",
    importpath = "github.com/rogpeppe/go-internal",
    sum = "h1:RR9dF3JtopPvtkroDZuVD7qquD0bnHlKSqaQhgwt8yk=",
    version = "v1.3.0",
)

go_repository(
    name = "com_github_russross_blackfriday_v2",
    importpath = "github.com/russross/blackfriday/v2",
    sum = "h1:JIOH55/0cWyOuilr9/qlrm0BSXldqnqwMsf35Ld67mk=",
    version = "v2.1.0",
)

go_repository(
    name = "com_github_sahilm_fuzzy",
    importpath = "github.com/sahilm/fuzzy",
    sum = "h1:FzWGaw2Opqyu+794ZQ9SYifWv2EIXpwP4q8dY1kDAwI=",
    version = "v0.1.0",
)

go_repository(
    name = "com_github_sergi_go_diff",
    importpath = "github.com/sergi/go-diff",
    sum = "h1:we8PVUC3FE2uYfodKH/nBHMSetSfHDR6scGdBi+erh0=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_shurcool_githubv4",
    importpath = "github.com/shurcooL/githubv4",
    sum = "h1:0/ecDXh/HTHRtSDSFnD2/Ta1yQ5J76ZspVY4u0/jGFk=",
    version = "v0.0.0-20200928013246-d292edc3691b",
)

go_repository(
    name = "com_github_shurcool_graphql",
    importpath = "github.com/shurcooL/graphql",
    sum = "h1:KikTa6HtAK8cS1qjvUvvq4QO21QnwC+EfvB+OAuZ/ZU=",
    version = "v0.0.0-20200928012149-18c5c3165e3a",
)

go_repository(
    name = "com_github_sirupsen_logrus",
    importpath = "github.com/sirupsen/logrus",
    sum = "h1:GL2rEmy6nsikmW0r8opw9JIRScdMF5hA8cOYLH7In1k=",
    version = "v1.4.1",
)

go_repository(
    name = "com_github_sourcegraph_jsonrpc2",
    importpath = "github.com/sourcegraph/jsonrpc2",
    sum = "h1:ohJHjZ+PcaLxDUjqk2NC3tIGsVa5bXThe1ZheSXOjuk=",
    version = "v0.1.0",
)

go_repository(
    name = "com_github_spf13_cobra",
    importpath = "github.com/spf13/cobra",
    sum = "h1:y+wJpx64xcgO1V+RcnwW0LEHxTKRi2ZDPSBjWnrg88Q=",
    version = "v1.4.0",
)

go_repository(
    name = "com_github_spf13_pflag",
    importpath = "github.com/spf13/pflag",
    sum = "h1:iy+VFUOCP1a+8yFto/drg2CJ5u0yRoB7fZw3DKv/JXA=",
    version = "v1.0.5",
)

go_repository(
    name = "com_github_stretchr_objx",
    importpath = "github.com/stretchr/objx",
    sum = "h1:2vfRuCMp5sSVIDSqO8oNnWJq7mPa6KVP3iPIwFBuy8A=",
    version = "v0.1.1",
)

go_repository(
    name = "com_github_stretchr_testify",
    importpath = "github.com/stretchr/testify",
    sum = "h1:5TQK59W5E3v0r2duFAb7P95B6hEeOyEnHRa8MjYSMTY=",
    version = "v1.7.1",
)

go_repository(
    name = "com_github_treilik_bubbleboxer",
    importpath = "github.com/treilik/bubbleboxer",
    sum = "h1:xA1dxzKGRDHTUiaj+p1MHWKtIMCpGIqir3IuSla74rw=",
    version = "v0.1.0",
)

go_repository(
    name = "com_github_xanzy_ssh_agent",
    importpath = "github.com/xanzy/ssh-agent",
    sum = "h1:wUMzuKtKilRgBAD1sUb8gOwwRr2FGoBVumcjoOACClI=",
    version = "v0.3.0",
)

go_repository(
    name = "com_github_xo_terminfo",
    importpath = "github.com/xo/terminfo",
    sum = "h1:QldyIu/L63oPpyvQmHgvgickp1Yw510KJOqX7H24mg8=",
    version = "v0.0.0-20210125001918-ca9a967f8778",
)

go_repository(
    name = "com_github_yuin_goldmark",
    importpath = "github.com/yuin/goldmark",
    sum = "h1:zNWRjYUW32G9KirMXYHQHVNFkXvMI7LpgNW2AgYAoIs=",
    version = "v1.4.4",
)

go_repository(
    name = "com_github_yuin_goldmark_emoji",
    importpath = "github.com/yuin/goldmark-emoji",
    sum = "h1:ctuWEyzGBwiucEqxzwe0SOYDXPAucOrE9NQC18Wa1os=",
    version = "v1.0.1",
)

go_repository(
    name = "com_google_cloud_go",
    importpath = "cloud.google.com/go",
    sum = "h1:Dg9iHVQfrhq82rUNu9ZxUDrJLaxFUe/HlCVaLyRruq8=",
    version = "v0.65.0",
)

go_repository(
    name = "com_google_cloud_go_bigquery",
    importpath = "cloud.google.com/go/bigquery",
    sum = "h1:PQcPefKFdaIzjQFbiyOgAqyx8q5djaE7x9Sqe712DPA=",
    version = "v1.8.0",
)

go_repository(
    name = "com_google_cloud_go_datastore",
    importpath = "cloud.google.com/go/datastore",
    sum = "h1:/May9ojXjRkPBNVrq+oWLqmWCkr4OU5uRY29bu0mRyQ=",
    version = "v1.1.0",
)

go_repository(
    name = "com_google_cloud_go_pubsub",
    importpath = "cloud.google.com/go/pubsub",
    sum = "h1:ukjixP1wl0LpnZ6LWtZJ0mX5tBmjp1f8Sqer8Z2OMUU=",
    version = "v1.3.1",
)

go_repository(
    name = "com_google_cloud_go_storage",
    importpath = "cloud.google.com/go/storage",
    sum = "h1:STgFzyU5/8miMl0//zKh2aQeTyeaUH3WN9bSUiJ09bA=",
    version = "v1.10.0",
)

go_repository(
    name = "com_shuralyov_dmitri_gpu_mtl",
    importpath = "dmitri.shuralyov.com/gpu/mtl",
    sum = "h1:VpgP7xuJadIUuKccphEpTJnWhS2jkQyMt6Y7pJCD7fY=",
    version = "v0.0.0-20190408044501-666a987793e9",
)

go_repository(
    name = "in_gopkg_check_v1",
    importpath = "gopkg.in/check.v1",
    sum = "h1:Hei/4ADfdWqJk1ZMxUNpqntNwaWcugrBjAiHlqqRiVk=",
    version = "v1.0.0-20201130134442-10cb98267c6c",
)

go_repository(
    name = "in_gopkg_errgo_v2",
    importpath = "gopkg.in/errgo.v2",
    sum = "h1:0vLT13EuvQ0hNvakwLuFZ/jYrLp5F3kcWHXdRggjCE8=",
    version = "v2.1.0",
)

go_repository(
    name = "in_gopkg_warnings_v0",
    importpath = "gopkg.in/warnings.v0",
    sum = "h1:wFXVbFY8DY5/xOe1ECiWdKCzZlxgshcYVNkBHstARME=",
    version = "v0.1.2",
)

go_repository(
    name = "in_gopkg_yaml_v2",
    importpath = "gopkg.in/yaml.v2",
    sum = "h1:D8xgwECY7CYvx+Y2n4sBz93Jn9JRvxdiyyo8CTfuKaY=",
    version = "v2.4.0",
)

go_repository(
    name = "in_gopkg_yaml_v3",
    importpath = "gopkg.in/yaml.v3",
    sum = "h1:h8qDotaEPuJATrMmW04NCwg7v22aHH28wwpauUhK9Oo=",
    version = "v3.0.0-20210107192922-496545a6307b",
)

go_repository(
    name = "io_opencensus_go",
    importpath = "go.opencensus.io",
    sum = "h1:LYy1Hy3MJdrCdMwwzxA/dRok4ejH+RwNGbuoD9fCjto=",
    version = "v0.22.4",
)

go_repository(
    name = "io_rsc_binaryregexp",
    importpath = "rsc.io/binaryregexp",
    sum = "h1:HfqmD5MEmC0zvwBuF187nq9mdnXjXsSivRiXN7SmRkE=",
    version = "v0.2.0",
)

go_repository(
    name = "io_rsc_quote_v3",
    importpath = "rsc.io/quote/v3",
    sum = "h1:9JKUTTIUgS6kzR9mK1YuGKv6Nl+DijDNIc0ghT58FaY=",
    version = "v3.1.0",
)

go_repository(
    name = "io_rsc_sampler",
    importpath = "rsc.io/sampler",
    sum = "h1:7uVkIFmeBqHfdjD+gZwtXXI+RODJ2Wc4O7MPEh/QiW4=",
    version = "v1.3.0",
)

go_repository(
    name = "org_golang_google_api",
    importpath = "google.golang.org/api",
    sum = "h1:yfrXXP61wVuLb0vBcG6qaOoIoqYEzOQS8jum51jkv2w=",
    version = "v0.30.0",
)

go_repository(
    name = "org_golang_google_appengine",
    importpath = "google.golang.org/appengine",
    sum = "h1:FZR1q0exgwxzPzp/aF+VccGrSfxfPpkBqjIIEq3ru6c=",
    version = "v1.6.7",
)

go_repository(
    name = "org_golang_google_genproto",
    importpath = "google.golang.org/genproto",
    sum = "h1:PDIOdWxZ8eRizhKa1AAvY53xsvLB1cWorMjslvY3VA8=",
    version = "v0.0.0-20200825200019-8632dd797987",
)

go_repository(
    name = "org_golang_google_grpc",
    importpath = "google.golang.org/grpc",
    sum = "h1:T7P4R73V3SSDPhH7WW7ATbfViLtmamH0DKrP3f9AuDI=",
    version = "v1.31.0",
)

go_repository(
    name = "org_golang_google_protobuf",
    importpath = "google.golang.org/protobuf",
    sum = "h1:SnqbnDw1V7RiZcXPx5MEeqPv2s79L9i7BJUlG/+RurQ=",
    version = "v1.27.1",
)

go_repository(
    name = "org_golang_x_crypto",
    importpath = "golang.org/x/crypto",
    sum = "h1:7I4JAnoQBe7ZtJcBaYHi5UtiO8tQHbUSXxL+pnGRANg=",
    version = "v0.0.0-20210921155107-089bfa567519",
)

go_repository(
    name = "org_golang_x_exp",
    importpath = "golang.org/x/exp",
    sum = "h1:QE6XYQK6naiK1EPAe1g/ILLxN5RBoH5xkJk3CqlMI/Y=",
    version = "v0.0.0-20200224162631-6cc2880d07d6",
)

go_repository(
    name = "org_golang_x_image",
    importpath = "golang.org/x/image",
    sum = "h1:+qEpEAPhDZ1o0x3tHzZTQDArnOixOzGD9HUJfcg0mb4=",
    version = "v0.0.0-20190802002840-cff245a6509b",
)

go_repository(
    name = "org_golang_x_lint",
    importpath = "golang.org/x/lint",
    sum = "h1:Wh+f8QHJXR411sJR8/vRBTZ7YapZaRvUcLFFJhusH0k=",
    version = "v0.0.0-20200302205851-738671d3881b",
)

go_repository(
    name = "org_golang_x_mobile",
    importpath = "golang.org/x/mobile",
    sum = "h1:4+4C/Iv2U4fMZBiMCc98MG1In4gJY5YRhtpDNeDeHWs=",
    version = "v0.0.0-20190719004257-d2bd2a29d028",
)

go_repository(
    name = "org_golang_x_mod",
    importpath = "golang.org/x/mod",
    sum = "h1:RM4zey1++hCTbCVQfnWeKs9/IEsaBLA8vTkd0WVtmH4=",
    version = "v0.3.0",
)

go_repository(
    name = "org_golang_x_net",
    importpath = "golang.org/x/net",
    sum = "h1:O7DYs+zxREGLKzKoMQrtrEacpb0ZVXA5rIwylE2Xchk=",
    version = "v0.0.0-20220127200216-cd36cc0744dd",
)

go_repository(
    name = "org_golang_x_oauth2",
    importpath = "golang.org/x/oauth2",
    sum = "h1:qfl7ob3DIEs3Ml9oLuPwY2N04gymzAW04WsUQHIClgM=",
    version = "v0.0.0-20220309155454-6242fa91716a",
)

go_repository(
    name = "org_golang_x_sync",
    importpath = "golang.org/x/sync",
    sum = "h1:5KslGYwFpkhGh+Q16bwMP3cOontH8FOep7tGV86Y7SQ=",
    version = "v0.0.0-20210220032951-036812b2e83c",
)

go_repository(
    name = "org_golang_x_sys",
    importpath = "golang.org/x/sys",
    sum = "h1:OH54vjqzRWmbJ62fjuhxy7AxFFgoHN0/DPc/UrL8cAs=",
    version = "v0.0.0-20220319134239-a9b59b0215f8",
)

go_repository(
    name = "org_golang_x_term",
    importpath = "golang.org/x/term",
    sum = "h1:JGgROgKl9N8DuW20oFS5gxc+lE67/N3FcwmBPMe7ArY=",
    version = "v0.0.0-20210927222741-03fcf44c2211",
)

go_repository(
    name = "org_golang_x_text",
    importpath = "golang.org/x/text",
    sum = "h1:olpwvP2KacW1ZWvsR7uQhoyTYvKAupfQrRGBFM352Gk=",
    version = "v0.3.7",
)

go_repository(
    name = "org_golang_x_time",
    importpath = "golang.org/x/time",
    sum = "h1:/5xXl8Y5W96D+TtHSlonuFqGHIWVuyCkGJLwGh9JJFs=",
    version = "v0.0.0-20191024005414-555d28b269f0",
)

go_repository(
    name = "org_golang_x_tools",
    importpath = "golang.org/x/tools",
    sum = "h1:W07d4xkoAUSNOkOzdzXCdFGxT7o2rW4q8M34tB2i//k=",
    version = "v0.0.0-20200825202427-b303f430e36d",
)

go_repository(
    name = "org_golang_x_xerrors",
    importpath = "golang.org/x/xerrors",
    sum = "h1:go1bK/D/BFZV2I8cIQd1NKEZ+0owSTG1fDTci4IqFcE=",
    version = "v0.0.0-20200804184101-5ec99f83aff1",
)

load("//:deps.bzl", "go_dependencies")

# gazelle:repository_macro deps.bzl%go_dependencies
go_dependencies()

go_rules_dependencies()

go_register_toolchains(version = "1.19.1")

gazelle_dependencies()
