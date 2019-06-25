package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	mySippMker "sippFileMker"
	"strconv"
	"sync"

	"github.com/Unknwon/goconfig"
)

type callConfig struct {
	local_ip        string
	remote_ip       string
	remote_port     string
	test_type       string
	insecure_invite string
	codec           string
	duration        string
	inter_time      string
	max_call        string
	simultaneous    string
	rate            string
	rate_period     string
	audio           string
	csv_file        string
}

func checkLsExists() {
    path, err := os.exec.LookPath("sipp")
    if err != nil {
        fmt.Printf("didn't find 'sipp' executable\n")
    } else {
        fmt.Printf("'sipp' executable is in '%s'\n", path)
    }
}

func main() {
	// fmt.Printf("This is a sipp test script configured by golang to test the sip protocol\nAuthor: yongchengLei\n")
	cfgFile := "SIPp.conf"
	x := configtest(cfgFile)
	csvCmd, pcapCmd, sippCmd := x.mkCmd()
	fmt.Printf("pcapCmd is: %v \ncsvCmd  is: %v \n\nsippCmd is: %v \n", pcapCmd, csvCmd, sippCmd)
	x.mkAllFile() // 开始生成sipp呼叫配置文件
	runCmd(pcapCmd, csvCmd, sippCmd)
}

func configtest(file string) callConfig {
	cfg, err := goconfig.LoadConfigFile(file)
	if err != nil {
		panic("error")
	}
	val, _ := cfg.GetSection("sipp")
	// fmt.Println(val)
	call := callConfig{
		val["local_ip"],
		val["remote_ip"],
		val["remote_port"],
		val["test_type"],
		val["insecure_invite"],
		val["codec"],
		val["duration"],
		val["inter_time"],
		val["max_call"],
		val["simultaneous"],
		val["rate"],
		val["rate_period"],
		val["audio"],
		val["csv_file"],
	}
	return call
}

func (c *callConfig) csvcopy() string {
	newfileName := "tmp/sipp-" + strconv.Itoa(os.Getpid()) + ".csv"
	return newfileName
}

func (c *callConfig) pscpcopy() string {
	pcapCode := c.codec
	newfileName := "sipp-" + pcapCode + "-" + strconv.Itoa(os.Getpid()) + ".pcap"
	return newfileName
}
func (c *callConfig) sippCmd() string {
	callFile := "tmp/sipp-" + strconv.Itoa(os.Getpid()) + ".xml"
	csvFile := c.csvcopy()
	oocFile := "tmp/ooc-" + strconv.Itoa(os.Getpid()) + ".xml"

	sippCmd := fmt.Sprintf("/usr/local/bin/sipp -i %v %v:%v -sf %v -inf %v -d %v -m %v -l %v -r %v -rp %v -oocsf %v", c.local_ip, c.remote_ip, c.remote_port, callFile, csvFile, c.duration, c.max_call, c.simultaneous, c.rate, c.rate_period, oocFile)
	// sipp 日志部分
	// sippCmd += fmt.Sprintf(" -trace_stat -stf tmp/sipp-statistics-%v.csv -fd 30 -periodic_rtd",strconv.Itoa(os.Getpid()))
	sippCmd += " -trace_stat -stf tmp/sipp-statistics-" + strconv.Itoa(os.Getpid()) + ".csv"
	sippCmd += " -trace_rtt"
	sippCmd += " -trace_screen"
	sippCmd += " -trace_err -error_file tmp/sipp-error-" + strconv.Itoa(os.Getpid()) + ".log"
	sippCmd += " -trace_error_codes"
	//sipp_cmd += "" -trace_msg -message_file tmp/sipp-message-%s.log -message_overwrite True' % pid
	//sipp_cmd += " -trace_shortmsg -shortmessage_file tmp/sipp-shortmsg-%s.log -shortmessage_overwrite True' % pid
	//sipp_cmd += " -trace_counts"
	//sipp_cmd += " -trace_calldebug"
	//sipp_cmd +=  -trace_logs"
	return sippCmd
}

func (c *callConfig) mkAllFile() {
	localDir, _ := os.Getwd()
	oocFile := localDir + "/tmp/ooc-" + strconv.Itoa(os.Getpid()) + ".xml"
	sipp_xml := localDir + "/tmp/sipp-" + strconv.Itoa(os.Getpid()) + ".xml"
	audio_file := ""
	if c.audio == "True" || c.audio == "true" {
		audio_file = "tmp/sipp-" + c.codec + "-" + strconv.Itoa(os.Getpid()) + ".pcap"
	}
	// fmt.Println(oocFile)
	args := []string{}
	mySippMker.MkScenario(oocFile, "out_of_call", "400", args) // 其他消息的处理场景 生成oocfile 。

	switch c.test_type {
	case "register":
		mySippMker.MkScenario(sipp_xml, c.test_type, c.inter_time, args) // 注册注销场景 生成注册注销时的xml
	case "call", "register_call":
		args = append(args, "codec", c.codec, "audio_file", audio_file, "insecure_invite", c.insecure_invite)
		mySippMker.MkScenario(sipp_xml, c.test_type, c.inter_time, args) // 呼叫、注册呼叫注销场景 ， 生成呼叫、注册注销时的xml
	default:
		fmt.Println("Test Type ERROR.....")
	}
}

func (c *callConfig) mkCmd() (csvCmd, pcapCmd, sippCmd string) {
	newcsv := c.csvcopy()
	if c.audio == "True" && c.test_type != "register" {
		newpcap := c.pscpcopy()
		// fmt.Println(newpcap)
		basePacpFile := fmt.Sprintf("demo-instruct.%v.pcap", c.codec)
		pcapCmd += fmt.Sprintf("cd tmp ; rm -rf %v ; ln -s ../pcap/%v %v ; cd ..", newpcap, basePacpFile, newpcap)
	}
	sippCmd = c.sippCmd()
	csvCmd += fmt.Sprintf("cp %v %v", c.csv_file, newcsv)
	// fmt.Printf("csvCmd  is: %v \npcapCmd is: %v \n", csvCmd, pcapCmd)
	return csvCmd, pcapCmd, sippCmd
}

func runCmd(pcapCmd, csvCmd, sippCmd string) {
	localDir, _ := os.Getwd()
	// if len(pcapCmd) == 0 {
	totalCmd := fmt.Sprintf("cd %v;%v;%v;%v", localDir, pcapCmd, csvCmd, sippCmd)
	cmd := exec.Command("/bin/sh", "-c", totalCmd)

	/*
	 获取标准输出和标准错误输出的两个管道
	*/
	var stdout, stderr []byte
	var errStdout, errStderr error
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	err := cmd.Start()
	if err != nil {
		log.Fatal("cmd.Start() failed with '%s'\n", err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		stdout, errStdout = copyAndCapture(os.Stdout, stdoutIn)
		wg.Done()
	}()

	stderr, errStderr = copyAndCapture(os.Stderr, stderrIn)

	wg.Wait()

	err = cmd.Wait()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	if errStdout != nil || errStderr != nil {
		log.Fatal("failed to capture stdout or stderr\n")
	}
	outStr, errStr := string(stdout), string(stderr)
	fmt.Printf("\nout:\n%s\nerr:\n%s\n", outStr, errStr)
}

func copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := w.Write(d)
			if err != nil {
				return out, err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
}
