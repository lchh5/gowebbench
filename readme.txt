Usage of gowebbench.exe:
  -c int
        clients 并发数
  -cfile string
        cookies save file path
	cookie文件的保存路径，内容支持 a=1&b=2|c=3&d=4 格式，当希望支持多组cookie进行处理时，可以用|进行格式，如希望支持模拟多个用户，则可以用
ebh_auth=aaaaa|ebh_auth=bbb
  -cookies string
        模拟的cookies值，可以直接命令行方式，多组cookie以|隔开
  -d string
        提交的post参数和值，格式为 a=b&c=d 这样的格式
  -dfile string
        支持将post参数保存到文件提交，此时传入保存的文件路径即可
  -f string
        此参数用于上传文件，输入需要上传的文件路径即可
  -fname string
        和-f组合，如果需要上传的文件以特定字段时可以指定
  -m string
        提交的方法，默认GET，支持GET和POST，当 参数 -d -f -dfile有值时，自动切换到POST，大部分情况下此值可以留空
  -t int
        bench time,if arg n set,it will be igorn
	测试的时间，以秒为单位
  -url string
        bench url 需要测试的URL地址，当 -url为空时，可以在命令最后面输入url

命令行示例：
gowebbench.exe -t 10 -c 200 -f e:/1.txt -d="a=b&c=1" -cfile="e:/c.txt" http://ss.ebh.net/bench.html 