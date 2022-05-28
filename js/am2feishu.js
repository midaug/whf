// alertmanager 转换 发送至飞书
// alertmanager的json格式可参考 json/alertmanager.json文件

// console.log("js run log ->  bodyStr: ", JSON.stringify(JSON.parse(bodyStr)))
var bodys = [];

var bodyObj = JSON.parse(bodyStr);

var alerts = bodyObj["alerts"];

var fields = ["labels", "annotations"];
for (var i in alerts) {
    var alert = alerts[i];
    var statusStr = alert["status"];
    var color = "red";
    if (statusStr === "resolved") {
        color = "green";
    }
    var time = dateFormat(new Date(alert["startsAt"]));
    var content = "";
    content = content + "**状态：**" + statusStr + "\n";
    content = content + "**时间：**" + time + "\n";

    for (var i in fields) {
        content = content + "====" + fields[i] + "====" + "\n";
        var fattr = alert[fields[i]];
        for (var key in fattr) {
            content = content + "**" + key + "：**" + fattr[key] + "\n";
        }
    }

    var newBody = {
        "msg_type": "interactive",
        "card": {
            "config": {
                "wide_screen_mode": true,
                "enable_forward": false
            },
            "elements": [
                {
                    "tag": "div",
                    "text": {
                        "content": content,
                        "tag": "lark_md"
                    }
                }
            ],
            "header": {
                "title": {
                    "content": title ? title : "告警通知",
                    "tag": "plain_text"
                },
                "template": color
            }
        }
    }
    bodys.push(JSON.stringify(newBody));
}

// go会读取body与contentType进行http请求，body是个数组字符串，影响发送次数
var body = JSON.stringify(bodys);
var contentType;