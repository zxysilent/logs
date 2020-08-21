## 简单 golang 日志记录库

日志结构
~~~ golang
// logger
type FishLogger struct {
	cons     bool          // 标准输出  默认 false
	callInfo bool          // 是否输出行号和文件名 默认 false
	maxAge   int           // 最大保留天数
	maxSize  int64         // 单个日志最大容量 默认 256MB
	size     int64         // 累计大小
	lpath    string        // 文件目录 完整路径 lpath=lname+lsuffix
	lname    string        // 文件名 无后缀
	lsuffix  string        // 文件后缀名 默认 .log
	created  string        // 文件创建日期
	level    logLevel      // 输出的日志等级
	list     *buffer       // 缓存
	listLock sync.Mutex    // 链表🔒
	lock     sync.Mutex    // logger🔒
	writer   *bufio.Writer // 缓存io 缓存到文件
	file     *os.File      // 日志文件
}
~~~

##使用示例
~~~ golang
import "github.com/zxysilent/logs"
func main() {
    // 退出时调用，确保日志写入文件中
    defer logs.Flush()
    // 设置日志输出等级
    // 开发环境下设置输出等级为DEBUG，线上环境设置为INFO
    logs.SetLevel(logs.DEBUG)
    // 设置输出调用信息
    logs.SetCallInfo(true)
    // 设置同时显示到控制台 
    // 默认只输出到文件
    logs.SetConsole(true)
    logs.Debug("Debug Logger")
	logs.Debugf("Debugf %s", "Logger")

	logs.Info("Info Logger")
	logs.Infof("Infof %s", "Logger")

	logs.Warn("Warn Logger")
	logs.Warnf("Warnf %s", "Logger")

	logs.Error("Error Logger")
	logs.Errorf("Errorf %s", "Logger")

  	logs.Fatal("Fatal Logger")
	logs.Fatalf("Fatalf %s", "Logger")
 }
 ~~~