
/********************************************************************
 created:    2022-11-27
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

export class Heartbeat
{
    public interval = 10   // 心跳间隔
    public timeoutId: any = null

    public clearTimeout() {
        if (this.timeoutId){
            clearTimeout(this.timeoutId)
            this.timeoutId = null
        }
    }
}