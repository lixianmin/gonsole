package client

/********************************************************************
created:    2023-11-26
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

// todo 感觉写各种兼容策略, 最终可能还不如重新写一遍golang版本的client来得快且实用

//type ClientSession1 struct {
//	serde        serde.Serde
//	onHandShaken func(bean *serde.JsonHandshake)
//
//	reconnectAction func()
//}
//
//func NewClientSession() *ClientSession1 {
//	var my = &ClientSession1{}
//	return my
//}
//
//func (my *ClientSession1) Connect(hostNameOrAddress string, port int, serde serde.Serde, onHandeShaken func(bean *serde.JsonHandshake)) {
//	_ = my.Close()
//	my.reconnectAction = func() {
//		var conn net.Conn
//		var err error
//		if len(tlsConfig) > 0 {
//			conn, err = tls.Dial("tcp", addr, tlsConfig[0])
//		} else {
//			conn, err = net.Dial("tcp", addr)
//		}
//
//		if err != nil {
//			return err
//		}
//
//		var link = intern.NewTcpLink(conn)
//		my.session = my.manager.NewSession(link).(road.ClientSession)
//		my.session.OnReceivedPacket(my.onReceivedPacketAtClient)
//	}
//}
//
//func (my *ClientSession1) Close() error {
//	return nil
//}
//
////public void Connect(string hostNameOrAddress, int port, ISerde serde, Action<JsonHandshake> onHandShaken = null)
////{
////Close();
////_reconnectAction = () =>
////{
////_serde = serde ?? throw new ArgumentNullException(nameof(serde));
////_onHandShaken = onHandShaken;
////
////var addressList = Dns.GetHostAddresses(hostNameOrAddress);
////if (!addressList.IsNullOrEmpty())
////{
////var address = addressList[0];
////_socket = new Socket(AddressFamily.InterNetwork, SocketType.Stream, ProtocolType.Tcp);
////_socket.Blocking = true;
////_socket.SetSocketOption(SocketOptionLevel.Tcp, SocketOptionName.NoDelay,
////true); // Disable the Nagle algorithm for this client
////_socket.Connect(address, port);
////
////_receiverThread = new ReceiverThread(_socket);
////}
////};
////
////_reconnectAction();
////}
