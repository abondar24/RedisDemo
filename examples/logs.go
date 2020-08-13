package examples

import "time"

const CommonKey = "common:"

func (cl *Client) RunLogs() {
	name1 := "ss1"
	msg1 := "error"

	name2 := "ss2"
	msg2 := "big error"

	name3 := "ss3"
	msg3 := "huge error"

	cl.logRecent(&name1, &msg1)
	cl.logRecent(&name2, &msg2)
	cl.logRecent(&name3, &msg3)
}

func (cl *Client) logRecent(name, msg *string) {
	message := time.Now().Local().String() + ":" + *msg
	destination := RecentKey + *name

	pipe := cl.client.Pipeline()
	pipe.LPush(*cl.ctx, destination, message)
	pipe.LTrim(*cl.ctx, destination, 0, 99)
	pipe.Exec(*cl.ctx)
}
