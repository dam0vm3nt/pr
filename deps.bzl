load("@bazel_gazelle//:deps.bzl", "go_repository")

def go_dependencies():
    go_repository(
        name = "com_github_evertras_bubble_table",
        importpath = "github.com/evertras/bubble-table",
        sum = "h1:JqmJw1/PaWJiaAJBnIgtoft7bLS+FVa23zOzv0uIHyY=",
        version = "v0.14.6",
    )
    go_repository(
        name = "com_github_muesli_cancelreader",
        importpath = "github.com/muesli/cancelreader",
        sum = "h1:SOpr+CfyVNce341kKqvbhhzQhBPyJRXQaCtn03Pae1Q=",
        version = "v0.2.0",
    )
