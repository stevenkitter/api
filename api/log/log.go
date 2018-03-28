package log

import (
	"fmt"
	log "github.com/cihub/seelog"
)

func SetupLog() {
	defer log.Flush()
	runWriter(consoleWriter)
	runWriter(rollingFileWriter)
	runWriter(smtpWriter)
}

func LogError(msg string) {
	defer log.Flush()
	log.Error("Error is %s", msg)
}

func LogWarn(msg string) {
	defer log.Flush()
	log.Warn("Hello from Seelog!")
}

func LogInfo(msg string) {
	defer log.Flush()
	log.Info("info is %s", msg)
}

func runWriter(writer func()) {
	writer()
	fmt.Println("---log------------seq----")
}

func consoleWriter() {
	testConfig := `
		<seelog>
			<outputs>
				<console />
			</outputs>
		</seelog>
		`
	logger, _ := log.LoggerFromConfigAsBytes([]byte(testConfig))
	log.UseLogger(logger)
	fmt.Println("Console writer")
	doLog()
}

func rollingFileWriter() {
	testConfig := `
<seelog>
	<outputs>
		<rollingfile type="size" filename="./log/roll.log" maxsize="1000" maxrolls="100" />
	</outputs>
</seelog>
`
	logger, _ := log.LoggerFromConfigAsBytes([]byte(testConfig))
	log.UseLogger(logger)
	fmt.Println("Rolling file writer")

	doLog()
}

func smtpWriter() {
	testConfig := `
 <seelog>
  <outputs>
   <smtp senderaddress="wangxu@julu666.com" sendername="Julu Manager" hostname="smtp.exmail.qq.com" hostport="995" username="wangxu@julu666.com" password="Julu2017">
    <recipient address="developer@julu666.com"/>
   </smtp>
  </outputs>
 </seelog>
 `
	logger, _ := log.LoggerFromConfigAsBytes([]byte(testConfig))
	log.UseLogger(logger)
	fmt.Println("SMTP writer is now sending emails to the specified recipients")
	doLog()
}

func doLog() {
	for i := 0; i < 5; i++ {
		log.Tracef("%d", i)
	}
}
