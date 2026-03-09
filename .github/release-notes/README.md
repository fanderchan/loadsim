# 发布说明

每个公开 tag 都必须有对应的发布说明文件：

```text
.github/release-notes/<tag>.md
```

例如：

```text
.github/release-notes/v0.3.0.md
```

固定结构：

```md
## [新增]
- ...

## [变更]
- ...

## [修复]
- ...

## [移除]
- ...
```

规则：

1. 四个章节必须按上面的顺序出现。
2. 如果某个章节没有内容，写 `- 无。`
3. 发布说明正文优先对比上一个正式版来撰写。
4. GitHub Release 会自动追加 compare 链接形式的“变更对比”。
5. 不要把 GitHub 自动生成的 release notes 当成最终发布正文。
