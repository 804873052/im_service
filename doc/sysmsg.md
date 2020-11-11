#系统消息内容格式:

1. 添加好友:

        v = {
            "sender":请求方用户id, 
            "receiver":接收方用户id, 
            "content":请求内容, 
            "timestamp":时间戳(秒)
        }
        op = {"add_friend_request":v}


2. 同意/拒绝添加好友:
 
        v = {
            "sender":请求方用户id, 
            "receiver":接收方用户id, 
            "status":答复状态(0.同意;1.拒绝), 
            "timestamp":时间戳(秒)
        }

        op = {"add_friend_reply":v}

3. 用户踢下线:
 
        v = {
            "token":请求token,
            "sender":请求方用户id, 
            "timestamp":时间戳(秒)
        }

        op = {"kick_user":v}
