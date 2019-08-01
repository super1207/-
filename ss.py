import requests
import re
import io
from PIL import Image
from pyzbar import pyzbar
import base64
import os

headers = {
'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.140 Safari/537.36 Edge/18.17763',
}

str1 = '''{
	"configs" : [
		{
			"remarks" : "",
			"id" : "CC5D18FE26581A603EEB406C7D454DCF",
'''
str2 = '''"server_udp_port" : 0,
			"protocol" : "origin",
			"protocolparam" : "",
			"obfs" : "plain",
			"obfsparam" : "",
			"remarks_base64" : "",
			"group" : "",
			"enable" : true,
			"udp_over_tcp" : false
		}
	],
	"index" : 0,
	"random" : false,
	"sysProxyMode" : 2,
	"shareOverLan" : false,
	"localPort" : 1080,
	"localAuthPassword" : "GBvf8VOc4QqqvgMXKZvv",
	"dnsServer" : "",
	"reconnectTimes" : 2,
	"randomAlgorithm" : 3,
	"randomInGroup" : false,
	"TTL" : 0,
	"connectTimeout" : 5,
	"proxyRuleMode" : 2,
	"proxyEnable" : false,
	"pacDirectGoProxy" : false,
	"proxyType" : 0,
	"proxyHost" : "",
	"proxyPort" : 0,
	"proxyAuthUser" : "",
	"proxyAuthPass" : "",
	"proxyUserAgent" : "",
	"authUser" : "",
	"authPass" : "",
	"autoBan" : false,
	"sameHostForSameTarget" : false,
	"keepVisitTime" : 180,
	"isHideTips" : true,
	"nodeFeedAutoUpdate" : true,
	"serverSubscribes" : [

	],
	"token" : {

	},
	"portMap" : {

	}
}'''
r = requests.get("https://ss.freeshadowsocks.biz",headers = headers)
pattern = re.compile(r'img/portfolio/ss[^r].*?\.png')
image_list = pattern.findall(r.content.decode('utf-8'))
image_list = ['https://ss.freeshadowsocks.biz/'+i for i in image_list]
print(image_list)
for url in image_list:
    ss64_list = pyzbar.decode(Image.open(io.BytesIO(requests.get(url,headers = headers).content)).convert('RGBA'), symbols=[pyzbar.ZBarSymbol.QRCODE])
    ss64 = ss64_list[0].data.decode('utf-8')[5:]
    ss = base64.b64decode(ss64).decode('utf-8')
    info = [i.split('@') for i in ss.split(":")]
    ssinfo = '''                        "server" : "'''+info[1][1]+'''",
                        "server_port" : '''+info[2][0]+''',
                        "password" : "'''+info[1][0]+'''",
                        "method" : "'''+info[0][0]+'''",'''
    with open('gui-config.json','w') as f:
        f.write(str1 + ssinfo + str2)
    print(str1 + ssinfo + str2)
    break
os.startfile('ShadowsocksR-dotnet2.0.exe') # start ShadowsocksR-dotnet2.0.exe process
