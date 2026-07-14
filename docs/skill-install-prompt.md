# Skill 安装 Prompt

```text
使用 octo-cli 将内置 Skills 安装到当前 Agent runtime。

1. 运行 `octo-cli version` 检查是否已安装。

2. 如果未安装，提供以下信息：
   - 项目地址：https://github.com/Mininglamp-OSS/octo-cli
   - 推荐安装：`npm install -g @mininglamp-oss/octo-cli`
   - Go 安装：`go install github.com/Mininglamp-OSS/octo-cli/cmd/octo-cli@latest`

3. 如果 octo-cli 版本低于 `0.7.0`，提示用户升级后再继续。

4. 判断当前 Agent runtime 及其 Skills 目录。

5. 执行：
   `octo-cli skills --install <skills目录>`

6. 检查安装后的 `SKILL.md`，确认安装成功。

7. 告知用户 octo-cli 版本、安装路径以及是否需要重启 runtime。
```
