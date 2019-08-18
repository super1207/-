import requests
import re
import json
import time
#import webbrowser
import telnetlib
import random
import sys
import colorama
from colorama import Fore, Back, Style
import os
import traceback
import win32ui
import win32con
import requests
from contextlib import closing
typ = ''

#文件下载器
def Down_load(file_url,file_path):
    headers = {"User-Agent":"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36"}
    with closing(requests.get(url = file_url,verify = False,headers=headers,stream=True)) as response:
        chunk_size = 1024 * 128  # 单次请求最大值
        try:
            content_size = int(response.headers['content-length'])  # 内容体总大小
        except:
            content_size = 999999999999
        data_count = 0
        with open(file_path, "wb") as file:
            for data in response.iter_content(chunk_size=chunk_size):
                file.write(data)
                data_count = data_count + len(data)
                now_jd = (data_count / content_size) * 100
                print("\r文件下载进度：%d%%(%d/%d) - %s" % (now_jd, data_count, content_size, file_path), end=" ")


def getProxyHttp(type):
    response = requests.get(url = 'https://raw.githubusercontent.com/fate0/proxylist/master/proxy.list')
    proxies_list = response.text.split('\n')
    http_proxies_list = [i for i in proxies_list if r'"'+type+'"' in i]
    random.shuffle(http_proxies_list)
    host = ""
    port = 0
    for proxy in http_proxies_list:
        proxy_json = json.loads(proxy)
        try:
            telnetlib.Telnet(proxy_json['host'],port=proxy_json['port'],timeout=3)
        except:
            pass
        else:
            host = proxy_json['host']
            port = proxy_json['port']
            break
    else:
        raise "代理获取失败"
    return host+":"+str(port)

def downbaidu(url,proxies={}):
    global typ
    while True:
        typ = input('输入要转换到的类型[word,pdf,ppt]:')
        typ = typ.lower()
        if typ == 'word':typ = 'doc'
        if typ in ['doc','pdf','ppt']:
            break
        print('输入格式错误,重新输入')
    print('正在获取授权...')
    r = requests.get('http://wenku.baiduvvv.com/doc/',proxies= proxies,verify = False).text
    sign = re.search(r'(name="sign" value=")(.*?)(" />)',r).groups()[1]
    t = re.search(r'(name="t" value=")(.*?)(" />)',r).groups()[1]
    # url = 'https://wenku.baidu.com/view/81ec2234580216fc700afd79.html?rec_flag=default&sxts=1566057150414'
    print('正在提交文库链接...')
    r = requests.get('http://wenku.baiduvvv.com/ds.php?url='+ url + '&type='+typ+ '&t='+ t +'&sign=' + sign,proxies= proxies,verify = False).text
    j = json.loads(r)
    s = j['s']
    f = j['f']
    h = j['h']
    # print(sign,t,j)
    urlt = s + '/wkc.php?url='+ url + '&type='+typ+'&t='+ t +'&sign='+ sign + '&f='+ f +'&h='+str(h)+'&btype=start&callback=callback2'
    print('开始转换...')
    r = requests.get(urlt,proxies= proxies,verify = False).content.decode('utf-8')
    while True:
        print('---')
        j = json.loads(re.findall(r'^\w+\((.*)\)$',r)[0])
        if j['code'] == 2:
            break
        elif j['code'] == -1:
            print('error',j['msg'])
            raise Exception(j['msg'])
        elif j['code'] == 1:
            urlt = s + '/wkc.php?url='+ url + '&type='+typ+'&t='+ t +'&sign='+ sign + '&f='+ f +'&h='+str(h)+'&btype=getProgress&callback=callback2'
            urlt = urlt + '&id='+j['id']
        elif j['code'] == 3:
            print('转换进度：',str(j['p'])+'%\r',end = '')
            sys.stdout.flush()
            r = requests.get(urlt,proxies= proxies,verify = False).content.decode('utf-8')
        else:
            print(r)
            raise Exception('未知错误！')

    ret = s + '/wkc.php?url='+ url + '&type='+typ+'&t='+ t +'&sign='+ sign + '&f='+ f +'&h='+str(h)+'&btype=down'
    print('转换进度：','转换成功!')
    print(ret)
    return ret

if __name__ == "__main__":
    while True:
        try:
            os.system('mode con: cols=80 lines=20')
            colorama.init()
            print('                           '+Back.GREEN + '百度文库下载器 SUPER1207' + Style.RESET_ALL)
            print('正在获取http代理...')
            proxies = {'http:':getProxyHttp('http')}
            print(proxies)
            while True:
                url = input('输入百度文库连接:').strip()
                downurl = downbaidu(url = url,proxies = proxies)
                if typ == 'doc':
                    typ = 'docx'
                name = str(random.randint(10000,99999))+'.'+ typ
                dlg = win32ui.CreateFileDialog(0, None, name, 0, typ + " File(*." + typ + ")|*.*||")
                if dlg.DoModal() == win32con.IDOK:
                    file_name = dlg.GetPathName()
                else:
                    print('取消下载')
                    continue
                print(file_name)
                print('正在下载...')
                Down_load(downurl,file_name)
                print('下载完成')
                # ebbrowser.open(downurl)
        except:
            print(traceback.format_exc())
            time.sleep(5)
