import requests
import bs4
import random
import json
import traceback
import os
import psutil
import winreg
import ctypes


def get_pid(name):
    arr = []
    for proc in psutil.process_iter(attrs=['pid', 'name']):
        if proc.info['name'] == 'Shadowsocks.exe':
            arr.append(proc.info['pid'])
    return arr
        
def kill(pid):
    try:
        import subprocess  
        subprocess.Popen("taskkill /F /T /PID %i"%pid , shell=True)  
    except:
        print('no process')
def disproxy():
    try:
        key = winreg.OpenKey(winreg.HKEY_CURRENT_USER,r"Software\Microsoft\Windows\CurrentVersion\Internet Settings",0, winreg.KEY_ALL_ACCESS)
        winreg.SetValueEx(key,'ProxyEnable',0,winreg.REG_DWORD,0)
        winreg.DeleteValue(key,'AutoConfigURL')
        winreg.CloseKey(key)
        internet_set_option = ctypes.windll.Wininet.InternetSetOptionW
        internet_set_option(0,39,0,0)
        internet_set_option(0,37,0,0)
    except:
        pass

def getssall():
    r = requests.get('https://blog.netimed.cn/ssr.html/comment-page-1')
    soup = bs4.BeautifulSoup(r.content.decode('utf-8'),features="html.parser")
    tbody = soup.select('#content > div > div > div.post_body > table > tbody > tr')
    l = random.randint(0,len(tbody) - 1)
    arr = []
    for i in tbody:
        objtd = i.select('td')
        mp = {"server":objtd[1].get_text(),"server_port":int(objtd[2].get_text()),"password":objtd[4].get_text(),"method":objtd[3].get_text(),"remarks" : ""}
        arr.append(mp)
    return {"configs" :arr,"index" : l,"global" : False,"enabled" : True,"shareOverLan" : False,"isDefault" : False,"localPort" : 1080}

def getpac():
    r = requests.get('https://cdn.jsdelivr.net/gh/cdlaimin/gfwlist2pac/gfwlist.pac')
    return r.content
if __name__ == "__main__":
    try:
        print('kill:Shadowsocks.exe')
        pidlist = get_pid('Shadowsocks.exe')
        for i in pidlist:
            kill(i)
        disproxy()
        print('config:pac.txt')
        with open('pac.txt','w+') as f:
            f.write(getpac().decode('utf-8'))
        print('config:gui-config.json')
        with open('gui-config.json','w+') as f:
            f.write(json.dumps(getssall()))
        print('start:shadowsocks.exe')
        os.popen('Shadowsocks.exe')
        print('all ok')
    except:
        traceback.print_exc()
        a = input('press any key to continue...')
