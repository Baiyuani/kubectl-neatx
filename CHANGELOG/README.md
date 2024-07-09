
## 发布操作

```shell
## 本地测试发布
goreleaser release --snapshot --clean

## 仅为给定的 GOOS/GOARCH 构建二进制文件
goreleaser build --single-target

## 发布到github Release
export GITHUB_TOKEN="YOUR_GH_TOKEN"
git tag -a v0.1.0 -m "First release"
git push origin v0.1.0
goreleaser release --release-notes CHANGELOG/CHANGELOG-1.0.0.md --clean
```

