package sippFileMker

// package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"io/ioutil"
	"log"
)

func main() {
	// Make_oocfile("tmp/", "1000")
	// MakeRegisterSec()
	args := []string{
		"codec", "ulaw", "audio_file", "this is audio_file", "insecure_invite", "True",
	}
	localDir, _ := os.Getwd()
	filename := fmt.Sprintf("%v/testfile", localDir)

	MkScenario(filename, "register", "50000", args)
}

func GetRealValue(cfg interface{}) map[string]string {
	// func GetRealValue(cfg interface{}) {
	cfgMap := make(map[string]string)
	val := reflect.ValueOf(cfg)
	// typ := val.Type()
	kd := val.Kind()
	if kd != reflect.Struct {
		fmt.Println("expect struct...")
		cfgMap["cfgstatus"] = "false"
		return cfgMap
	}
	for i := 0; i < val.NumField(); i++ {
		f := val.Field(i)
		a1 := fmt.Sprint(val.Type().Field(i).Name)
		a2 := fmt.Sprintf("%v", f)
		fmt.Printf("\t%v-->%v\n", a1, a2)
		cfgMap[a1] = a2
	}
	return cfgMap
}

func MakeUacSdpBody(codec string) string {
	payload_type := 0
	encoding_name := ""
	switch codec {
	case "alaw":
		payload_type = 8
		encoding_name = "PCMA"
	case "ulaw":
		payload_type = 0
		encoding_name = "PCMU"
	case "g729":
		payload_type = 18
		encoding_name = "G729"
	case "g723":
		payload_type = 4
		encoding_name = "G723"
	case "g726":
		payload_type = 2
		encoding_name = "G726"
	}
	body := "v=0\n"
	body += "o=[field1] [pid][call_number] 8[pid][call_number]8 IN IP[local_ip_type] [local_ip]\n"
	body += "s=SIPp Media\n"
	body += "i=Media Data\n"
	body += "c=IN IP[media_ip_type] [media_ip]\n"
	body += "t=0 0\n"
	body += fmt.Sprintf("m=audio [media_port] RTP/AVP %d 101\n", payload_type)
	body += fmt.Sprintf("a=rtpmap:%d %v/8000\n", payload_type, encoding_name)
	body += "a=rtpmap:101 telephone-event/8000\n"
	body += "a=fmtp:101 0-15,16\n"
	body += "a=sendrecv\n"
	return body
}

func make_scenario_start_end(inter_time string) (start, end string) {
	/*SIPp 场景起始、结束部分
	参数：
		inter_time：每轮呼叫的时间间隔
	返回值：返回元组(起始部分,结束部分)
	*/
	start = "<?xml version=\"1.0\" encoding=\"ISO-8859-1\" ?>\n"
	start += "<!DOCTYPE scenario SYSTEM \"sipp.dtd\">\n>"
	start += "<scenario name=\"SIPp scenario\">\n"

	end = fmt.Sprintf("<timewait milliseconds=\"%v\"/>\n", inter_time)
	end += "<ResponseTimeRepartition value=\"10, 20, 30, 40, 50, 100, 150, 200\"/>\n"
	end += "<CallLengthRepartition value=\"10, 50, 100, 500, 1000, 5000, 10000\"/>\n"
	end += "</scenario>\n"
	return start, end
}

// func makeRegSec(){}

// }

// func UacSendRequest(cfgMap map[string]string, method, auth string, kwargs ...map[string]string)(request string) {
func UacSendRequest(method string, args []string) (request string) {
	// func UacSendRequest(cfg interface{}) {
	/*
				UAC 发送的 SIP 请求部分 (request string)
		        参数：
		            method：SIP 方法，UNREGISTER 表示注销
		            auth：SIP 请求是否携带验证
		            **kwargs：其他参数，支持 retrans、start_rtd、rtd、Expires、SDP
		        返回值：返回 UAC SIP 请求字符串
	*/
	request += "<send"

	kwargs := make(map[string]string)
	for i := 0; i < len(args); i += 2 {
		// fmt.Println(x[i])
		kwargs[args[i]] = args[i+1]
	}

	if k, ok := kwargs["retrans"]; ok {
		request += fmt.Sprintf(" retrans=\"%v\"", k)
	}
	if k, ok := kwargs["start_rtd"]; ok {
		request += fmt.Sprintf(" start_rtd=\"%v\"", k)
	}
	if k, ok := kwargs["rtd"]; ok {
		request += fmt.Sprintf(" rtd=\"%v\"", k)
	}
	request += ">\n"
	request += "<![CDATA[\n"
	if method == "REGISTER" {
		request += fmt.Sprintf("%s sip:[remote_ip]:[remote_port] SIP/2.0\n", method)
	} else {
		request += fmt.Sprintf("%v sip:[field3]@[remote_ip]:[remote_port] SIP/2.0\n", method)
	}
	request += "Via: SIP/2.0/[transport] [local_ip]:[local_port];branch=[branch]\n"
	request += "Max-Forwards: 70\n"
	request += "Contact: <sip:[field1]@[local_ip]:[local_port];transport=[transport]>\n"
	if strings.Contains(method, "REGISTER") {
		request += "To: [field0]<sip:[field1]@[remote_ip]:[remote_port]>"
	} else {
		request += "To: <sip:[field1]@[remote_ip]:[remote_port]>"
	}
	if method == "ACK" || method == "BYE" {
		request += "[peer_tag_param]"
	}
	request += "\n"

	if strings.Contains(method, "REGISTER") {
		request += "From: \"[field0]\"<sip:[field1]@[remote_ip]:[remote_port]>;tag=[pid]SIPpTagRegister[call_number]\n"
	} else {
		request += "From: \"[field0]\"<sip:[field1]@[remote_ip]:[remote_port]>;tag=[pid]SIPpTagInvite[call_number]\n"
	}
	request += "Call-ID: [call_id]\n"
	if strings.Contains(method, "REGISTER") {
		request += "CSeq: [cseq] REGISTER\n"
	} else {
		request += fmt.Sprintf("CSeq: [cseq] %v\n", method)
	}
	if k, ok := kwargs["Expires"]; ok {
		request += fmt.Sprintf("Expires: %v\n", k)
	}
	if method == "INVITE" {
		request += "Content-Type: application/sdp\n"
	}
	request += "User-Agent: SIPp\n"
	request += "Subject: Call Performance Test made by LeiYongCheng\n"

	if k, ok := kwargs["auth"]; ok {
		if k == "True" {
			request += "[field2]\n"
		}
	}
	if method == "INVITE" {
		request += "Content-Length: [len]\n"
	} else {
		request += "Content-Length: 0\n"
	}
	if k, ok := kwargs["SDP"]; ok {
		request += fmt.Sprintf("%v", k)
	}
	request += "]]>\n"
	request += "</send>\n"

	return request
}

func UacRecvStatus(method string, args []string) string {
	/*
				UAC 接收响应
		        参数：
		            method：SIP 方法，如 401/200
		            auth：增加 auth="true"
		            optional：是否可选
		            crlf_sec：统计界面增加换行
		        返回值：返回 UAC 接收响应的字符串
	*/

	// 将可变参数转成key:value 形式
	kwargs := make(map[string]string)
	for i := 0; i < len(args); i += 2 {
		// fmt.Println(x[i])
		kwargs[args[i]] = args[i+1]
	}
	status := fmt.Sprintf("<recv response=\"%v\"", method)
	if _, ok := kwargs["auth"]; ok {
		status += " auth=\"true\""
	}
	if _, ok := kwargs["optional"]; ok {
		status += " optional=\"true\""
	}
	if _, ok := kwargs["crlf"]; ok {
		status += " crlf=\"true\""
	}
	if k, ok := kwargs["start_rtd"]; ok {
		status += fmt.Sprintf(" start_rtd=\"%v\"", k)
	}
	if k, ok := kwargs["rtd"]; ok {
		status += fmt.Sprintf(" rtd=\"%v\"", k)
	}
	status += "></recv>\n"
	return status

}

func MakeRegisterSec() (registerSec, unregisterSec string) {
	/*
		SIPp 场景：注册、注销部分
			返回值：返回元组(注册部分,注销部分)
	*/
	registerSecSlice1 := []string{"auth", "False", "Expires", "3600", "retrans", "500", "start_rtd", "register"}
	registerSecSlice2 := []string{"auth", "True", "rtd", "register"}
	registerSecSlice3 := []string{"auth", "True", "Expires", "3600", "retrans", "500"}
	registerSecSlice4 := []string{"crlf", "True"}

	unregisterSecSlice1 := []string{"auth", "True", "Expires", "0", "retrans", "500", "start_rtd", "unregister"}
	unregisterSecSlice2 := []string{"auth", "True", "rtd", "unregister"}
	unregisterSecSlice3 := []string{"auth", "True", "Expires", "0", "retrans", "500"}
	unregisterSecSlice4 := []string{"crlf", "True"}

	registerSec += UacSendRequest("REGISTER", registerSecSlice1)
	registerSec += UacRecvStatus("401", registerSecSlice2)
	registerSec += UacSendRequest("REGISTER", registerSecSlice3)
	registerSec += UacRecvStatus("200", registerSecSlice4)

	unregisterSec += UacSendRequest("REGISTER", unregisterSecSlice1)
	unregisterSec += UacRecvStatus("401", unregisterSecSlice2)
	unregisterSec += UacSendRequest("REGISTER", unregisterSecSlice3)
	unregisterSec += UacRecvStatus("200", unregisterSecSlice4)

	return registerSec, unregisterSec
}

func MakeCallSec(insecure_invite, codec, audio_file string) string {
	registerSecSlice1Sdp := MakeUacSdpBody(codec)
	registerSecSlice1 := []string{"auth", "False", "retrans", "500", "start_rtd", "invite", "SDP", registerSecSlice1Sdp}
	call_sec := UacSendRequest("INVITE", registerSecSlice1)
	UacRecvStatusSlice1 := []string{"optional", "True", "rtd", "invite"}
	if insecure_invite == "True" {
		// UacRecvStatusSlice1 := []string{"optional", "True", "rtd", "invite"}
		UacSendRequestSlice1 := []string{"auth", "False", "retrans", "500", "start_rtd", "bye"}
		call_sec += UacRecvStatus("100", UacRecvStatusSlice1)
		call_sec += UacRecvStatus("180", UacRecvStatusSlice1)
		call_sec += UacRecvStatus("183", UacRecvStatusSlice1)
		call_sec += UacRecvStatus("200", []string{"rtd", "invite"})
		call_sec += UacSendRequest("ACK", []string{"auth", "False"})
		call_sec += sipp_pause(audio_file)
		call_sec += UacSendRequest("BYE", UacSendRequestSlice1)
		call_sec += UacRecvStatus("200", []string{"crlf", "True", "rtd", "bye"})

	} else {
		call_sec += UacRecvStatus("401", []string{"auth", "True", "rtd", "invite"})
		call_sec += UacSendRequest("ACK", []string{"auth", "False"})
		call_sec += UacSendRequest("INVITE", []string{"auth", "True", "retrans", "500", "start_rtd", "invite", "SDP", registerSecSlice1Sdp})
		call_sec += UacRecvStatus("100", UacRecvStatusSlice1)
		call_sec += UacRecvStatus("180", UacRecvStatusSlice1)
		call_sec += UacRecvStatus("183", UacRecvStatusSlice1)
		call_sec += UacRecvStatus("200", []string{"rtd", "invite"})
		call_sec += UacSendRequest("ACK", []string{"auth", "True"})
		call_sec += sipp_pause(audio_file)
		call_sec += UacSendRequest("BYE", []string{"auth", "True", "retrans", "500", "start_rtd", "bye"})
		call_sec += UacRecvStatus("200", []string{"crlf", "True", "rtd", "bye"})
	}
	return call_sec
}

func sipp_pause(audio_file string) (action string) {
	/*SIPp 的呼叫暂停部分：
	  	注册与注销的暂停部分（不带语音流）
	  	INVITE 建立通话后的语音传输部分
	  参数：
	  	audio_file：携带的语音流文件名
	  返回值：返回暂停部分的字符串
	*/
	if audio_file != "" {
		action = fmt.Sprintf("<nop><action><exec play_pcap_audio=\"%v\"/></action></nop>\n", audio_file)
		action += "<pause/>\n"
	} else {
		action = "<pause/>\n"
	}
	return action
}

func uac_send_status(method string, args []string) (status string) {
	status_map := map[string]string{
		"100": "100 Trying",
		"180": "180 Ringing",
		"183": "183 Session Progress",
		"200": "200 OK",
	}
	status += fmt.Sprintf("<send>\n<![CDATA[\nSIP/2.0 %v\n", status_map[method])
	status += "[last_Via:]\n[last_From:]\n[last_To:]\n[last_Call-ID:]\n[last_CSeq:]\n"
	status += "Server: SIPp\n"
	status += "Contact: <sip:[local_ip]:[local_port];transport=[transport]>\n"
	kwargs := make(map[string]string)
	for i := 0; i < len(args); i += 2 {
		// fmt.Println(x[i])
		kwargs[args[i]] = args[i+1]
	}

	if _, ok := kwargs["SDP"]; ok {
		status += "Content-Length: [len]\n"
	} else {
		status += "Content-Length: 0\n"
	}
	if k, ok := kwargs["SDP"]; ok {
		status += k
	} 
	status += "]]>\n</send>\n"

	return status
}

// 生成场景部分。。。。。。

func regUnreg(interTime string) string {
	/*
		注册注销场景

		参数：
			inter_time：每轮测试的时间间隔

		返回值：返回场景字符串

	*/
	scenario, end := make_scenario_start_end(interTime)
	register_sec, unregister_sec := MakeRegisterSec()
	scenario += register_sec
	scenario += sipp_pause("")
	scenario += unregister_sec
	scenario += end
	return scenario
}

func regCallUnreg(interTime, codec, audio_file, insecure_invite string) string {
	/*
			SIPp 注册呼叫注销场景
		        参数：
		            inter_time：每轮测试的时间间隔
		            codec：使用的语音编码
		            audio_file：通话建立后使用的语音流
		            insecure_invite：是否验证 INVITE 消息
				返回值：返回场景字符串
	*/
	scenario, end := make_scenario_start_end(interTime)
	register_sec, unregister_sec := MakeRegisterSec()
	scenario += register_sec
	scenario += MakeCallSec(insecure_invite, codec, audio_file)
	scenario += unregister_sec
	scenario += end
	return scenario
}

func call(interTime, codec, audio_file, insecure_invite string) string {
	/*
			SIPp 呼叫场景
		        参数：
		            inter_time：每轮测试的时间间隔
		            codec：使用的语音编码
		            audio_file：通话建立后使用的语音流
		            insecure_invite：是否验证 INVITE 消息
		            返回值：返回场景字符串
	*/
	scenario, end := make_scenario_start_end(interTime)
	scenario += MakeCallSec(insecure_invite, codec, audio_file)
	scenario += end
	return scenario
}

func out_of_call(inter_time string) string {
	/*
		意外消息的处理场景：Out-of-call UAS
	*/
	scenario, end := make_scenario_start_end(inter_time)
	scenario += "<recv request=\".*\" regexp_match=\"true\"></recv>\n"
	scenario += uac_send_status("200",[]string{})
	scenario += end
	return scenario
}

func MkScenario(fileName, testType, interTime string, args []string) {
	// sipXmlFileName := fmt.Sprintf("%vsipp-%v.xml", dir, pid)
	// sipCsvFileName := fmt.Sprintf("%vsipp-%v.xml", dir, pid)
	// oocFileBytes := out_of_call("400")

	kwargsSlice := []string{}
	kwargsSliceStr := ""
	kwargs := make(map[string]string)

	if len(args) >= 2 {
		for i := 0; i < len(args); i += 2 {
			// fmt.Println(args[i], args[i+1])
			kwargs[args[i]] = args[i+1]
		}
		for k, _ := range kwargs {
			kwargsSlice = append(kwargsSlice, k)
		}
		kwargsSliceStr = strings.Join(kwargsSlice, " ")
		// fmt.Println(kwargsSliceStr)
	}

	switch testType {
	case "out_of_call":
		oocFileBytes := out_of_call("400") // 其他消息的处理场景 获取 oocbytes
		err := ioutil.WriteFile(fileName, []byte(oocFileBytes), 0666)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("其他消息的处理场景")
		// fmt.Println(oocFileBytes)
	case "register":
		registerBytes := regUnreg(interTime) // 注册注销场景 获取bytes
		err := ioutil.WriteFile(fileName, []byte(registerBytes), 0666)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("注册注销场景")
		// fmt.Println(registerBytes)
	case "call":
		if !strings.Contains(kwargsSliceStr, "codec") || !strings.Contains(kwargsSliceStr, "audio_file") || !strings.Contains(kwargsSliceStr, "insecure_invite") {
			fmt.Println("Use call or reg_call_unreg scenario, you must use codec&audio_file&insecure_invite parameters!")
		} else {
			callBytes := call(interTime, kwargs["codec"], kwargs["audio_file"], kwargs["insecure_invite"])
			err := ioutil.WriteFile(fileName, []byte(callBytes), 0666)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("呼叫场景..")
			// fmt.Println(callBytes)
		}
	case "register_call":
		if !strings.Contains(kwargsSliceStr, "codec") || !strings.Contains(kwargsSliceStr, "audio_file") || !strings.Contains(kwargsSliceStr, "insecure_invite") {
			fmt.Println("Use call or reg_call_unreg scenario, you must use codec&audio_file&insecure_invite parameters!")
		} else {
			registerBytes := regCallUnreg(interTime, kwargs["codec"], kwargs["audio_file"], kwargs["insecure_invite"])
			err := ioutil.WriteFile(fileName, []byte(registerBytes), 0666)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("注册呼叫场景。。。")
			// fmt.Println(registerBytes)
		}
	}
}