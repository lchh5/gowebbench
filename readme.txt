Usage of gowebbench.exe:
  -c int
        clients ������
  -cfile string
        cookies save file path
	cookie�ļ��ı���·��������֧�� a=1&b=2|c=3&d=4 ��ʽ����ϣ��֧�ֶ���cookie���д���ʱ��������|���и�ʽ����ϣ��֧��ģ�����û����������
ebh_auth=aaaaa|ebh_auth=bbb
  -cookies string
        ģ���cookiesֵ������ֱ�������з�ʽ������cookie��|����
  -d string
        �ύ��post������ֵ����ʽΪ a=b&c=d �����ĸ�ʽ
  -dfile string
        ֧�ֽ�post�������浽�ļ��ύ����ʱ���뱣����ļ�·������
  -f string
        �˲��������ϴ��ļ���������Ҫ�ϴ����ļ�·������
  -fname string
        ��-f��ϣ������Ҫ�ϴ����ļ����ض��ֶ�ʱ����ָ��
  -m string
        �ύ�ķ�����Ĭ��GET��֧��GET��POST���� ���� -d -f -dfile��ֵʱ���Զ��л���POST���󲿷�����´�ֵ��������
  -t int
        bench time,if arg n set,it will be igorn
	���Ե�ʱ�䣬����Ϊ��λ
  -url string
        bench url ��Ҫ���Ե�URL��ַ���� -urlΪ��ʱ���������������������url

������ʾ����
gowebbench.exe -t 10 -c 200 -f e:/1.txt -d="a=b&c=1" -cfile="e:/c.txt" http://ss.ebh.net/bench.html 