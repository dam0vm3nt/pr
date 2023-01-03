load("@bazel_gazelle//:deps.bzl", "go_repository")

def go_dependencies():
    go_repository(
        name = "com_github_99designs_gqlgen",
        importpath = "github.com/99designs/gqlgen",
        sum = "h1:yczvlwMsfcVu/JtejqfrLwXuSP0yZFhmcss3caEvHw8=",
        version = "v0.17.2",
    )
    go_repository(
        name = "com_github_agnivade_levenshtein",
        importpath = "github.com/agnivade/levenshtein",
        sum = "h1:QY8M92nrzkmr798gCo3kmMyqXFzdQVpxLlGPRBij0P8=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_alexflint_go_arg",
        importpath = "github.com/alexflint/go-arg",
        sum = "h1:lDWZAXxpAnZUq4qwb86p/3rIJJ2Li81EoMbTMujhVa0=",
        version = "v1.4.2",
    )
    go_repository(
        name = "com_github_alexflint_go_scalar",
        importpath = "github.com/alexflint/go-scalar",
        sum = "h1:NGupf1XV/Xb04wXskDFzS0KWOLH632W/EO4fAFi+A70=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_andreyvit_diff",
        importpath = "github.com/andreyvit/diff",
        sum = "h1:bvNMNQO63//z+xNgfBlViaCIJKLlCJ6/fmUseuG0wVQ=",
        version = "v0.0.0-20170406064948-c7f18ee00883",
    )
    go_repository(
        name = "com_github_arbovm_levenshtein",
        importpath = "github.com/arbovm/levenshtein",
        sum = "h1:jfIu9sQUG6Ig+0+Ap1h4unLjW6YQJpKZVmUzxsD4E/Q=",
        version = "v0.0.0-20160628152529-48b4e1c0c4d0",
    )
    go_repository(
        name = "com_github_bradleyjkemp_cupaloy_v2",
        importpath = "github.com/bradleyjkemp/cupaloy/v2",
        sum = "h1:knToPYa2xtfg42U3I6punFEjaGFKWQRXJwj0JTv4mTs=",
        version = "v2.6.0",
    )
    go_repository(
        name = "com_github_dgryski_trifles",
        importpath = "github.com/dgryski/trifles",
        sum = "h1:fRzb/w+pyskVMQ+UbP35JkH8yB7MYb4q/qhBarqZE6g=",
        version = "v0.0.0-20200323201526-dd97f9abfb48",
    )

    go_repository(
        name = "com_github_evertras_bubble_table",
        importpath = "github.com/evertras/bubble-table",
        sum = "h1:JqmJw1/PaWJiaAJBnIgtoft7bLS+FVa23zOzv0uIHyY=",
        version = "v0.14.6",
    )
    go_repository(
        name = "com_github_kevinmbeaulieu_eq_go",
        importpath = "github.com/kevinmbeaulieu/eq-go",
        sum = "h1:AQgYHURDOmnVJ62jnEk0W/7yFKEn+Lv8RHN6t7mB0Zo=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_khan_genqlient",
        importpath = "github.com/Khan/genqlient",
        sum = "h1:TMZJ+tl/BpbmGyIBiXzKzUftDhw4ZWxQZ+1ydn0gyII=",
        version = "v0.5.0",
    )
    go_repository(
        name = "com_github_logrusorgru_aurora_v3",
        importpath = "github.com/logrusorgru/aurora/v3",
        sum = "h1:R6zcoZZbvVcGMvDCKo45A9U/lzYyzl5NfYIvznmDfE4=",
        version = "v3.0.0",
    )
    go_repository(
        name = "com_github_matryer_moq",
        importpath = "github.com/matryer/moq",
        sum = "h1:Q06vEqnBYjjfx5KKgHfYRKE/lvlRu+Nj+xodG4YdHnU=",
        version = "v0.2.3",
    )
    go_repository(
        name = "com_github_mitchellh_mapstructure",
        importpath = "github.com/mitchellh/mapstructure",
        sum = "h1:f/MjBEBDLttYCGfRaKBbKSRVF5aV2O6fnBpzknuE3jU=",
        version = "v1.2.3",
    )

    go_repository(
        name = "com_github_muesli_cancelreader",
        importpath = "github.com/muesli/cancelreader",
        sum = "h1:SOpr+CfyVNce341kKqvbhhzQhBPyJRXQaCtn03Pae1Q=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_shurcool_sanitized_anchor_name",
        importpath = "github.com/shurcooL/sanitized_anchor_name",
        sum = "h1:PdmoCO6wvbs+7yrJyMORt4/BmY5IYyJwS/kOiWx8mHo=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_urfave_cli_v2",
        importpath = "github.com/urfave/cli/v2",
        sum = "h1:qph92Y649prgesehzOrQjdWyxFOp/QVM+6imKHad91M=",
        version = "v2.3.0",
    )
    go_repository(
        name = "com_github_vektah_gqlparser_v2",
        importpath = "github.com/vektah/gqlparser/v2",
        sum = "h1:C02NsyEsL4TXJB7ndonqTfuQOL4XPIu0aAWugdmTgmc=",
        version = "v2.4.5",
    )
