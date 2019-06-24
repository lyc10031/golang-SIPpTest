# golang-SIPpTest
文件说明：

           clean.sh：清除历史数据脚本。

           SIPp.conf ：测试配置文件。 

           sipp_user ：性能测试所需.csv文件存放路径。 

           SIPpRunner ：SIPp测试脚本

           pacp： 语音文件存放目录

           tmp：临时文件、运行日志存放目录。

使用说明：输入命令： 

            直接执行二进制文件SIPpRunner  进行测试。

            修改SIPp.conf 中的测试类型，test_type = call  可以进行三种测试，

            call对应SIPp 呼叫测试、register对应sip账号注册注销测试、register_call对应sip账号注册呼叫测试。

配置文件中其他选项说明：

            local_ip = 本地地址
            remote_ip = 呼叫地址
            remote_port = 呼叫端口
            test_type = 测试类型
            insecure_invite = 是否验证 INVITE
            codec = 音频编码
            duration = 通话时长（ms）
            inter_time = 呼叫间隔时间（ms）
            max_call = 总呼叫数
            simultaneous = 并发呼叫数
            rate = 呼叫速率
            rate_period = 呼叫周期
            csv_file = 用户配置文件
