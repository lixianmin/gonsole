var V=Object.defineProperty;var Z=(o,t,e)=>t in o?V(o,t,{enumerable:!0,configurable:!0,writable:!0,value:e}):o[t]=e;var r=(o,t,e)=>(Z(o,typeof t!="symbol"?t+"":t,e),e);import{h as Q,d as tt,a as et,c as it,b as O,w as nt,v as st,u as ot,i as rt,e as F,f as N,F as at,o as lt,s as ht,g as ct}from"./vendor.181eaea1.js";const ut=function(){const t=document.createElement("link").relList;if(t&&t.supports&&t.supports("modulepreload"))return;for(const n of document.querySelectorAll('link[rel="modulepreload"]'))i(n);new MutationObserver(n=>{for(const s of n)if(s.type==="childList")for(const h of s.addedNodes)h.tagName==="LINK"&&h.rel==="modulepreload"&&i(h)}).observe(document,{childList:!0,subtree:!0});function e(n){const s={};return n.integrity&&(s.integrity=n.integrity),n.referrerpolicy&&(s.referrerPolicy=n.referrerpolicy),n.crossorigin==="use-credentials"?s.credentials="include":n.crossorigin==="anonymous"?s.credentials="omit":s.credentials="same-origin",s}function i(n){if(n.ep)return;n.ep=!0;const s=e(n);fetch(n.href,s)}};ut();class B{static blockCopy(t,e,i,n,s){for(let h=0;h<s;h++)i[n++]=t[e++]}}var q=(o=>(o[o.Begin=0]="Begin",o[o.Current=1]="Current",o[o.End=2]="End",o))(q||{});const U=class{constructor(t){r(this,"initialIndex",0);r(this,"dirtyBytes",0);r(this,"position",0);r(this,"length",0);r(this,"capacity",0);r(this,"buffer");if(t<0)throw new Error(`capacity=${t}`);this.capacity=t,this.buffer=new Uint8Array(t)}setCapacity(t){if(t!=this.capacity&&t!=this.buffer.length){let e;if(t!=0){e=new Uint8Array(t);for(let i=0;i<this.length;i++)e[i]=this.buffer[i];this.buffer=e}this.dirtyBytes=0,this.capacity=t}}getCapacity(){return this.capacity-this.initialIndex}setPosition(t){if(t<0)throw new Error(`position=${t}`);this.position=this.initialIndex+t}getPosition(){return this.position-this.initialIndex}setLength(t){if(t>this.capacity)throw new Error("length can not be greater than capacity");if(t<0||t+this.initialIndex>U.maxCapacity)throw new Error("out of range");let e=t+this.initialIndex;e>this.length?this.expand(e):e<this.length&&(this.dirtyBytes+=this.length-e),this.length=e,this.position>this.length&&(this.position=this.length)}getLength(){return this.length-this.initialIndex}expand(t){if(t>this.capacity){let e=t;e<32?e=32:e<this.capacity<<1&&(e=this.capacity<<1),this.setCapacity(e)}else if(this.dirtyBytes>0){for(let e=0;e<this.dirtyBytes;e++){let i=e+this.length;this.buffer[i]=0}this.dirtyBytes=0}}readByte(){return this.position>=this.length?-1:this.buffer[this.position++]}read(t,e,i){if(e<0||i<0)throw new Error(`offset=${e}, count=${i}`);if(this.buffer.length-e<i)throw new Error("the size of buffer is less than offset + count");return this.position>=this.length||i==0?0:(this.position>=this.length-i&&(i=this.length-this.position),B.blockCopy(this.buffer,this.position,t,e,i),this.position+=i,i)}write(t,e,i){if(e<0||i<0)throw new Error(`offset=${e}, count=${i}`);if(t.byteLength-e<i)throw new Error("the size of the buffer is less than offset + count");this.position>this.length-i&&this.expand(this.position+i),B.blockCopy(t,e,this.buffer,this.position,i),this.position+=i,this.position>=this.length&&(this.length=this.position)}seek(t,e){if(t>U.maxCapacity)throw new Error("offset is out of range");let i;switch(e){case 0:if(t<0)throw new Error("attempted to seek before start of OctetsSteam");i=this.initialIndex;break;case 1:i=this.position;break;case 2:i=this.length;break;default:throw new Error("invalid SeekOrigin")}if(i+=t,i<this.initialIndex)throw new Error("attempted to seek before start of OctetsStream");return this.position=i,this.position}tidy(){const t=this.length-this.position;B.blockCopy(this.buffer,this.position,this.buffer,0,t),this.setPosition(0),this.setLength(t)}toString(){return`dirtyBytes=${this.dirtyBytes}, position=${this.position}, length=${this.length}, capacity=${this.capacity}, buffer=${this.buffer}`}};let G=U;r(G,"maxCapacity",2147483647);class H{constructor(t,e){r(this,"type");r(this,"body");this.type=t,this.body=e}static encode(t,e){const i=e?e.length:0,n=4,s=new Uint8Array(n+i);let h=0;return s[h++]=t&255,s[h++]=i>>16&255,s[h++]=i>>8&255,s[h++]=i&255,e&&B.blockCopy(e,0,s,h,i),s}static decode(t){const e=[];for(;t.getLength()-t.getPosition()>=4;){const n=t.readByte(),s=t.readByte()<<16|t.readByte()<<8|t.readByte()>>>0;if(s<0)throw new Error(`type=${n}, length = ${s}, stream=${t.toString()}`);if(t.getLength()<s){t.seek(-4,q.Current);break}const h=new Uint8Array(s);s>0&&t.read(h,0,s);let d=new H(n,h);e.push(d)}return t.tidy(),e}}var A=(o=>(o[o.Handshake=1]="Handshake",o[o.HandshakeAck=2]="HandshakeAck",o[o.Heartbeat=3]="Heartbeat",o[o.Data=4]="Data",o[o.Kick=5]="Kick",o))(A||{});function K(o){const t=new ArrayBuffer(o.length*3),e=new Uint8Array(t);let i=0;for(let s=0;s<o.length;s++){const h=o.charCodeAt(s);let d;h<=127?d=[h]:h<=2047?d=[192|h>>6,128|h&63]:d=[224|h>>12,128|(h&4032)>>6,128|h&63];for(let p=0;p<d.length;p++)e[i]=d[p],++i}const n=new Uint8Array(i);return B.blockCopy(e,0,n,0,i),n}function D(o){const t=new Uint8Array(o),e=[];let i=0,n=0;const s=t.length;for(;i<s;)t[i]<128?(n=t[i],i+=1):t[i]<224?(n=((t[i]&63)<<6)+(t[i+1]&63),i+=2):(n=((t[i]&15)<<12)+((t[i+1]&63)<<6)+(t[i+2]&63),i+=3),e.push(n);return dt(e)}function dt(o){const e=[];for(let i=0;i<o.length;i+=32768)e.push(String.fromCharCode.apply(null,o.slice(i,i+32768)));return e.join("")}var T=(o=>(o[o.Request=0]="Request",o[o.Notify=1]="Notify",o[o.Response=2]="Response",o[o.Push=3]="Push",o[o.Count=4]="Count",o))(T||{});const b=class{constructor(t,e,i,n,s){r(this,"id");r(this,"type");r(this,"compressRoute");r(this,"route");r(this,"body");this.id=t,this.type=e,this.compressRoute=i,this.route=n,this.body=s}static encode(t,e,i,n,s){const h=b.hasId(e)?b.calculateMsgIdBytes(t):0;let d=b.MSG_FLAG_BYTES+h;if(b.hasRoute(e)){if(i){if(typeof n!="number")throw new Error("error flag for number route!");d+=b.MSG_ROUTE_CODE_BYTES}else if(d+=b.MSG_ROUTE_LEN_BYTES,n){if(n=K(n),n.length>255)throw new Error("route maxlength is overflow");d+=n.length}}s&&(d+=s.length);const p=new Uint8Array(d);let w=0;return w=b.encodeMsgFlag(e,i,p,w),b.hasId(e)&&(w=b.encodeMsgId(t,p,w)),b.hasRoute(e)&&(w=b.encodeMsgRoute(i,n,p,w)),s!=null&&(w=b.encodeMsgBody(s,p,w)),p}static decode(t){const e=new Uint8Array(t),i=e.length||e.byteLength;let n=0,s=0,h="";const d=e[n++],p=d&b.MSG_COMPRESS_ROUTE_MASK,w=d>>1&b.MSG_TYPE_MASK;if(b.hasId(w)){let g=e[n],L=0;do g=e[n],s=s+(g&127)*Math.pow(2,7*L),n++,L++;while(g>=128)}if(b.hasRoute(w))if(p)h=(e[n++]<<8|e[n++]).toString();else{const g=e[n++];if(g){let L=new Uint8Array(g);B.blockCopy(e,n,L,0,g),h=D(h)}else h="";n+=g}const I=i-n,k=new Uint8Array(I);return B.blockCopy(e,n,k,0,I),new b(s,w,p,h,k)}static calculateMsgIdBytes(t){let e=0;do e+=1,t>>=7;while(t>0);return e}static encodeMsgFlag(t,e,i,n){if(!b.isValid(t))throw new Error("unknown message type: "+t);return i[n]=t<<1|(e?1:0),n+b.MSG_FLAG_BYTES}static encodeMsgId(t,e,i){do{let n=t%128;const s=Math.floor(t/128);s!==0&&(n=n+128),e[i++]=n,t=s}while(t!==0);return i}static encodeMsgRoute(t,e,i,n){if(t){if(e>b.MSG_ROUTE_CODE_MAX)throw new Error("route number is overflow");i[n++]=e>>8&255,i[n++]=e&255}else e?(i[n++]=e.length&255,B.blockCopy(e,0,i,n,e.length),n+=e.length):i[n++]=0;return n}static encodeMsgBody(t,e,i){return B.blockCopy(t,0,e,i,t.length),i+t.length}static hasId(t){return t===T.Request||t===T.Response}static hasRoute(t){return t===T.Request||t===T.Notify||t===T.Push}static isValid(t){return t>=T.Request&&t<T.Count}};let x=b;r(x,"MSG_FLAG_BYTES",1),r(x,"MSG_ROUTE_CODE_BYTES",2),r(x,"MSG_ID_MAX_BYTES",5),r(x,"MSG_ROUTE_LEN_BYTES",1),r(x,"MSG_ROUTE_CODE_MAX",65535),r(x,"MSG_COMPRESS_ROUTE_MASK",1),r(x,"MSG_TYPE_MASK",7);class ft{constructor(){r(this,"handleHeartBeat",t=>{!this.heartbeatInterval||(this.heartbeatTimeoutId&&(clearTimeout(this.heartbeatTimeoutId),this.heartbeatTimeoutId=null),!this.heartbeatId&&(this.heartbeatId=setTimeout(()=>{this.heartbeatId=null;const e=H.encode(A.Heartbeat);this.send(e),this.nextHeartbeatTimeout=Date.now()+this.heartbeatTimeout,this.heartbeatTimeoutId=setTimeout(this.heartbeatTimeoutCb,this.heartbeatTimeout)},this.heartbeatInterval)))});r(this,"handleHandshake",t=>{let e=JSON.parse(D(t));const i=501;if(e.code===i){this.emit("error","client version not fullfill");return}const n=200;if(e.code!==n){this.emit("error","handshake fail");return}this.handshakeInit(e);const s=H.encode(A.HandshakeAck);this.send(s)});r(this,"handleData",t=>{let e=t;this.decode&&(e=this.decode(e)),this.processMessage(e)});r(this,"handleKick",t=>{t=JSON.parse(D(t)),this.emit("onKick",t)});r(this,"socket",null);r(this,"buffer",new G(8));r(this,"useCrypto",!1);r(this,"encode");r(this,"decode");r(this,"requestIdGenerator",0);r(this,"reconnectUrl","");r(this,"reconnect",!1);r(this,"reconnectTimer");r(this,"reconnectAttempts",0);r(this,"reconnectionDelay",5e3);r(this,"handshakeBuffer",{sys:{type:"js-websocket",version:"0.0.1",rsa:{}},user:{}});r(this,"pushHandlers",new Map);r(this,"handlers",new Map);r(this,"routeMap",new Map);r(this,"callbacks",new Map);r(this,"abbrs",{});r(this,"dict",{});r(this,"heartbeatInterval",0);r(this,"heartbeatTimeout",0);r(this,"nextHeartbeatTimeout",0);r(this,"heartbeatTimeoutId");r(this,"heartbeatId");r(this,"handshakeCallback")}on(t,e){this.pushHandlers[t]=e}emit(t,e=""){const i=this.pushHandlers[t];i!=null&&i(e)}processPackages(t){for(let e=0;e<t.length;e++){const i=t[e],n=this.handlers[i.type];n!=null&&n(i.body)}}defaultDecode(t){const e=x.decode(t);if(!(e.id>0&&(e.route=this.routeMap[e.id],this.routeMap.delete(e.id),!e.route)))return e.body=this.decompose(e),e}decompose(t){let e=t.route;if(t.compressRoute){if(!this.abbrs[e])return{};e=t.route=this.abbrs[e]}return JSON.parse(D(t.body))}reset(){this.reconnect=!1,this.reconnectionDelay=1e3*5,this.reconnectAttempts=0,clearTimeout(this.reconnectTimer)}initData(t){if(!(!t||!t.sys)&&(this.dict=t.sys.dict,this.dict)){this.abbrs={};for(const e in this.dict)this.abbrs[this.dict[e]]=e}}handshakeInit(t){t.sys&&t.sys.heartbeat?(this.heartbeatInterval=t.sys.heartbeat*1e3,this.heartbeatTimeout=this.heartbeatInterval*2):(this.heartbeatInterval=0,this.heartbeatTimeout=0),this.initData(t),typeof this.handshakeCallback=="function"&&this.handshakeCallback(t.user)}processMessage(t){if(t.id){const e=this.callbacks[t.id];this.callbacks.delete(t.id),typeof e=="function"&&e(t.body)}else{const e=this.pushHandlers[t.route];typeof e!="undefined"&&e(t.body)}}heartbeatTimeoutCb(){const t=this.nextHeartbeatTimeout-Date.now();t>100?this.heartbeatTimeoutId=setTimeout(this.heartbeatTimeoutCb,t):(console.error("server heartbeat timeout"),this.emit("heartbeat timeout"),this.disconnect())}send(t){this.socket!=null?this.socket.send(t.buffer):console.log("socket = null")}sendMessage(t,e,i){let n=i;this.encode&&(n=this.encode(t,e,i));const s=H.encode(A.Data,n);this.send(s)}connectInner(t,e,i){console.log("connect to: "+e),t=t||{};const n=10,s=t.maxReconnectAttempts||n;this.reconnectUrl=e;const h=k=>{console.log("onopen",k),this.reconnect&&this.emit("reconnect"),this.reset();const S=H.encode(A.Handshake,K(JSON.stringify(this.handshakeBuffer)));this.send(S),i!=null&&i()},d=k=>{let S=new Uint8Array(k.data),g=this.buffer;g.write(S,0,S.length),g.setPosition(0),this.processPackages(H.decode(g)),this.heartbeatTimeout&&(this.nextHeartbeatTimeout=Date.now()+this.heartbeatTimeout)},p=k=>{this.emit("io-error",k),console.error("socket error: ",k)},w=k=>{this.emit("close",k),this.emit("disconnect",k),console.log("socket close: ",k),t.reconnect&&this.reconnectAttempts<s&&(this.reconnect=!0,this.reconnectAttempts++,this.reconnectTimer=setTimeout(()=>{this.connectInner(t,this.reconnectUrl,i)},this.reconnectionDelay),this.reconnectionDelay*=2)};let I=new WebSocket(e);I.binaryType="arraybuffer",I.onopen=h,I.onmessage=d,I.onerror=p,I.onclose=w,this.socket=I}connect(t,e){this.handshakeCallback=t.handshakeCallback,this.encode=t.encode||this.defaultEncode,this.decode=t.decode||this.defaultDecode,this.handshakeBuffer.user=t.user,this.handlers[A.Heartbeat]=this.handleHeartBeat,this.handlers[A.Handshake]=this.handleHandshake,this.handlers[A.Data]=this.handleData,this.handlers[A.Kick]=this.handleKick,this.connectInner(t,t.url,e)}defaultEncode(t,e,i){const n=t!=0?T.Request:T.Notify;i=K(JSON.stringify(i));let s=!1;return this.dict&&this.dict[e]&&(e=this.dict[e],s=!0),x.encode(t,n,s,e,i)}disconnect(){this.socket!=null&&(this.socket.close(),console.log("disconnect"),this.socket=null),this.heartbeatId&&(clearTimeout(this.heartbeatId),this.heartbeatId=null),this.heartbeatTimeoutId&&(clearTimeout(this.heartbeatTimeoutId),this.heartbeatTimeoutId=null)}request(t,e,i){let n=++this.requestIdGenerator;this.sendMessage(n,t,e),this.callbacks[n]=i,this.routeMap[n]=t}notify(t,e){e=e||{},this.sendMessage(0,t,e)}}function bt(o){const t=document.getElementById("mainPanel");if(t==null)return;let e=t.scrollTop>t.scrollHeight-t.clientHeight-1;t.appendChild(o),e&&(t.scrollTop=t.scrollHeight-t.clientHeight)}function v(o){const t=document.createElement("div");t.innerHTML=o,bt(t)}function _(){v("<br>")}function C(o){v("["+Q(new Date).format("HH:mm:ss.S")+"] "+o)}class pt{constructor(){r(this,"currentIndex",-1);r(this,"list",[]);const t="list",e=localStorage.getItem(t);if(e){const i=JSON.parse(e);i&&(this.list=i,this.currentIndex=this.list.length)}window.onunload=i=>{const n=this.list.slice(-100);localStorage.setItem(t,JSON.stringify(n))}}add(t){if(t!=null&&t!=""){const e=this.list,i=e.length;i==0||e[i-1]!==t?this.currentIndex=e.push(t):this.currentIndex=e.length}}getHistories(){return this.list}getHistory(t){return t>=0&&t<this.list.length?this.list[t]:""}move(t){if(t!=0){let e=this.currentIndex+t;if(e>=0&&e<this.list.length)return this.currentIndex=e,this.list[e]}return""}toString(){return`currentIndex=${this.currentIndex}, list=[${this.list}]`}}class gt{constructor(){r(this,"autoLoginLimit",0);r(this,"websocketPath","");r(this,"urlRoot","");r(this,"title","");r(this,"body","")}loadData(t){console.log(t),this.autoLoginLimit=t.autoLoginLimit,this.websocketPath=t.websocketPath,this.urlRoot=t.urlRoot,this.title=t.title,this.body=t.body}getAutoLoginLimit(){return this.autoLoginLimit}getWebsocketUrl(t){return`${document.location.protocol==="https:"?"wss:":"ws:"}//${t}/${this.websocketPath}`}getUrlRoot(){return this.urlRoot}getTitle(){return this.title}getBody(){return this.body}}const mt=O("div",{id:"mainPanel"},null,-1),yt={id:"inputBoxDiv"},wt=["onKeydown"],kt=tt({setup(o){let t=`${document.location.host}${document.title}`,e=`${document.location.protocol}//${t}`,i="",n="",s=!1,h=new gt,d=new pt,p=new ft,w="";et.get(e+"/web_config").then(a=>{h.loadData(a.data),console.log(w);let l=h.getWebsocketUrl(t);p.connect({url:l},()=>{console.log("websocket connected")}),p.on("disconnect",()=>{C("<b> disconnected from server </b>")}),document.title=h.getTitle(),v(h.getBody()),_(),p.on("console.html",I),p.on("console.default",k)}),window.onload=()=>{const a=document.getElementById("inputBox");!a||(a.focus(),document.onkeydown=function(l){if(l.key==="Enter"&&document.activeElement!==a&&a)return a.focus(),!1})};function I(a){C("<b>server\u54CD\u5E94\uFF1A</b>"+a.data),_()}function k(a){const l=JSON.stringify(a);C("<b>server\u54CD\u5E94\uFF1A</b>"+l),_()}function S(a,l,c){const u=JSON.stringify(l);C("<b>client\u8BF7\u6C42\uFF1A</b>"),v(u),_(),p.request(a,l,c)}function g(a){switch(a.op){case"log.list":X(a.data);break;case"history":Y(a.data);break;case"html":I(a);break;case"empty":break;default:k(a)}}function L(a){let l=a.target.value.trim();if(l!==""){if(a.target.value="",l.startsWith("!")){const f=parseInt(l.substring(1))-1;isNaN(f)||(l=d.getHistory(f))}let u=l.split(/\s+/),m=u.length;const y=u[0];if(y==="help"){const f={command:y+" "+e};S("console.command",f,g),d.add(l)}else if(m>=2&&(y==="sub"||y==="unsub")){const f={topic:u[1]},E="console."+y;S(E,f,g),d.add(l)}else if(m>=2&&y==="auth")n=u[1],s=!0,a.target.type="password",C(l+"<br/> <h3>\u8BF7\u8F93\u5165\u5BC6\u7801\uFF1A</h3><br/>"),d.add(l);else if(s&&m>=1){s=!1,a.target.type="text";const f=y;if(J(n,f),localStorage){const E="autoLoginUser",M={username:n,password:f,expireTime:new Date().getTime()+h.getAutoLoginLimit()},$=JSON.stringify(M);localStorage.setItem(E,$)}}else{const f={command:u.join(" ")};S("console.command",f,g),d.add(l)}}else C("");const c=document.getElementById("mainPanel");c&&(c.scrollTop=c.scrollHeight-c.clientHeight)}function J(a,l){const c="hey pet!",u=ht.exports.sha256.hmac(c,l),m={command:"auth "+a+" "+u};S("console.command",m,g)}function Y(a){const l=d.getHistories(),c=l.length,u=new Array(c);for(let y=0;y<c;y++)u[y]="<li>"+l[y]+"</li>";let m="<b>\u5386\u53F2\u547D\u4EE4\u5217\u8868\uFF1A</b> <br/> count:&nbsp;"+c+"<br/><ol>"+u.join("")+"</ol>";C(m),_()}function z(a){const l=a.target.value.trim();if(l.length>0){const c={head:l};p.request("console.hint",c,u=>{const m=u.names,y=u.notes,f=m.length;if(f>0&&(a.target.value=W(m),f>1)){const E=new Array(f);for(let R=0;R<f;R++)E[R]=`<tr> <td>${R+1}</td> <td>${m[R]}</td> <td>${y[R]}</td> </tr>`;const $="<table> <tr> <th></th> <th>Name</th> <th>Note</th> </tr>"+E.join("")+"</table>";C($),_()}})}}function j(a){a.key;const l=a.key=="ArrowUp"?-1:1,c=d.move(l);if(c!=""){let u=a.target;u.value=c,setTimeout(()=>{let m=c.length;u.setSelectionRange(m,m),u.focus()},0)}}function W(a){if(a.length<2)return a.join();let l=a[0];for(let c=1;c<a.length;c++)for(let u=l.length;u>0&&l!==a[c].substring(0,u);u--)l=l.substring(0,u-1);return l}function X(a){const l=a.logFiles,c=l.length,u=new Array(c);let m=0;for(let f=0;f<c;f++){const E=l[f];m+=E.size;let M=P(E.size);u[f]=`<tr> <td>${f+1}</td> <td>${M}</td> <td> <div class="tips"><a href="${e}/${E.path}">${E.path}</a> <span class="tips_text">${E.sample}</span>
                                <input type="button" class="copy_button" onclick="copyToClipboard('${E.path}')" value="\u590D\u5236"/>
                                </div></td> <td>${E.mod_time}</td> </tr>`}let y="<b>\u65E5\u5FD7\u6587\u4EF6\u5217\u8868\uFF1A</b> <br> count:&nbsp;"+c+"<br>total:&nbsp;&nbsp;"+P(m)+"<br>";y+="<table> <tr> <th></th> <th>Size</th> <th>Name</th> <th>Modified Time</th> </tr>"+u.join("")+"</table>",C(y),_()}function P(a){return a<1024?a+"B":a<1048576?(a/1024).toFixed(1)+"K":(a/1048576).toFixed(1)+"M"}return(a,l)=>(lt(),it(at,null,[mt,O("div",yt,[nt(O("input",{id:"inputBox",ref:"mainPanel","onUpdate:modelValue":l[0]||(l[0]=c=>rt(i)?i.value=c:i=c),placeholder:"Tab\u8865\u5168\u547D\u4EE4, Enter\u6267\u884C\u547D\u4EE4",onKeydown:[F(N(L,["prevent"]),["enter"]),F(N(z,["prevent"]),["tab"]),F(N(j,["prevent"]),["up","down"])]},null,40,wt),[[st,ot(i)]])])],64))}});ct(kt).mount("#app");
