load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test")
load("@bazel_gazelle//:def.bzl", "gazelle")

# gazelle:prefix github.com/rc-postcard/rc-postcard
gazelle(name = "gazelle")

go_library(
    name = "rc-postcard_lib",
    srcs = [
        "api.go",
        "api_postcards.go",
        "auth_middleware.go",
        "handlers.go",
        "main.go",
        "postgres.go",
    ],
    embedsrcs = [
        "static/back-of-4x6-postcard-1.html",
        "static/favicon.ico",
        "static/home.html",
        "static/script.js",
        "static/styles.css",
    ],
    importpath = "github.com/rc-postcard/rc-postcard",
    visibility = ["//visibility:private"],
    deps = [
        "//lob",
        "@com_github_google_uuid//:go_default_library",
        "@com_github_jackc_pgx_v4//stdlib:go_default_library",
        "@org_golang_x_oauth2//:go_default_library",
    ],
)

go_binary(
    name = "rc-postcard",
    embed = [":rc-postcard_lib"],
    visibility = ["//visibility:public"],
)

go_test(
    name = "rc-postcard_test",
    srcs = ["main_test.go"],
    embed = [":rc-postcard_lib"],
)
