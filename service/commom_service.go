package service

/*
关于错误处理：
1. 接口的成功调用与否取决4于Http的状态码，封装在BaseApi的OK，Fail，ServerFail方法中
2. 错误情况分为:
	a) 参数校验错误[Validation Error]		Service层返回error, 在Api层调用error.Error()方法获取组装进ResponseJson.Msg  e.g. 密码长度不够
	b) 业务冲突错误[Business Conflict Error] Service层返回error e.g. 用户名不唯一
*/
