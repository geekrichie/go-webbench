# go-webbench
go-webbench

Webbench是一个在linux下使用的非常简单的网站压测工具。通过使用go中的协程来并发请求访问url来测试网站的负载
灵感主要来源于C实现的webbench项目
## 依赖
   go 1.14及以上
   以下的版本暂没有进行测试
## 使用：
项目根目录下执行go build 生成二进制文件(之后打算集成makefile）
 <pre>
  go build 
  ./webench http://host:port/path/124.html</pre>
## 选项：
  待补充
