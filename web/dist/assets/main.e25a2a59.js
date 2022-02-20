var X=Object.defineProperty;var V=(o,t,e)=>t in o?X(o,t,{enumerable:!0,configurable:!0,writable:!0,value:e}):o[t]=e;var r=(o,t,e)=>(V(o,typeof t!="symbol"?t+"":t,e),e);import{h as Z,s as Q,d as tt,c as et,a as $,w as it,v as nt,u as st,i as ot,b as U,e as N,F as rt,o as at,f as ht}from"./vendor.a0254286.js";const lt=function(){const t=document.createElement("link").relList;if(t&&t.supports&&t.supports("modulepreload"))return;for(const n of document.querySelectorAll('link[rel="modulepreload"]'))i(n);new MutationObserver(n=>{for(const s of n)if(s.type==="childList")for(const h of s.addedNodes)h.tagName==="LINK"&&h.rel==="modulepreload"&&i(h)}).observe(document,{childList:!0,subtree:!0});function e(n){const s={};return n.integrity&&(s.integrity=n.integrity),n.referrerpolicy&&(s.referrerPolicy=n.referrerpolicy),n.crossorigin==="use-credentials"?s.credentials="include":n.crossorigin==="anonymous"?s.credentials="omit":s.credentials="same-origin",s}function i(n){if(n.ep)return;n.ep=!0;const s=e(n);fetch(n.href,s)}};lt();class B{static blockCopy(t,e,i,n,s){for(let h=0;h<s;h++)i[n++]=t[e++]}}var K=(o=>(o[o.Begin=0]="Begin",o[o.Current=1]="Current",o[o.End=2]="End",o))(K||{});const R=class{constructor(t){r(this,"initialIndex",0);r(this,"dirtyBytes",0);r(this,"position",0);r(this,"length",0);r(this,"capacity",0);r(this,"buffer");if(t<0)throw new Error(`capacity=${t}`);this.capacity=t,this.buffer=new Uint8Array(t)}setCapacity(t){if(t!=this.capacity&&t!=this.buffer.length){let e;if(t!=0){e=new Uint8Array(t);for(let i=0;i<this.length;i++)e[i]=this.buffer[i];this.buffer=e}this.dirtyBytes=0,this.capacity=t}}getCapacity(){return this.capacity-this.initialIndex}setPosition(t){if(t<0)throw new Error(`position=${t}`);this.position=this.initialIndex+t}getPosition(){return this.position-this.initialIndex}setLength(t){if(t>this.capacity)throw new Error("length can not be greater than capacity");if(t<0||t+this.initialIndex>R.maxCapacity)throw new Error("out of range");let e=t+this.initialIndex;e>this.length?this.expand(e):e<this.length&&(this.dirtyBytes+=this.length-e),this.length=e,this.position>this.length&&(this.position=this.length)}getLength(){return this.length-this.initialIndex}expand(t){if(t>this.capacity){let e=t;e<32?e=32:e<this.capacity<<1&&(e=this.capacity<<1),this.setCapacity(e)}else if(this.dirtyBytes>0){for(let e=0;e<this.dirtyBytes;e++){let i=e+this.length;this.buffer[i]=0}this.dirtyBytes=0}}readByte(){return this.position>=this.length?-1:this.buffer[this.position++]}read(t,e,i){if(e<0||i<0)throw new Error(`offset=${e}, count=${i}`);if(this.buffer.length-e<i)throw new Error("the size of buffer is less than offset + count");return this.position>=this.length||i==0?0:(this.position>=this.length-i&&(i=this.length-this.position),B.blockCopy(this.buffer,this.position,t,e,i),this.position+=i,i)}write(t,e,i){if(e<0||i<0)throw new Error(`offset=${e}, count=${i}`);if(t.byteLength-e<i)throw new Error("the size of the buffer is less than offset + count");this.position>this.length-i&&this.expand(this.position+i),B.blockCopy(t,e,this.buffer,this.position,i),this.position+=i,this.position>=this.length&&(this.length=this.position)}seek(t,e){if(t>R.maxCapacity)throw new Error("offset is out of range");let i;switch(e){case 0:if(t<0)throw new Error("attempted to seek before start of OctetsSteam");i=this.initialIndex;break;case 1:i=this.position;break;case 2:i=this.length;break;default:throw new Error("invalid SeekOrigin")}if(i+=t,i<this.initialIndex)throw new Error("attempted to seek before start of OctetsStream");return this.position=i,this.position}tidy(){const t=this.length-this.position;B.blockCopy(this.buffer,this.position,this.buffer,0,t),this.setPosition(0),this.setLength(t)}toString(){return`dirtyBytes=${this.dirtyBytes}, position=${this.position}, length=${this.length}, capacity=${this.capacity}, buffer=${this.buffer}`}};let F=R;r(F,"maxCapacity",2147483647);class _{constructor(t,e){r(this,"type");r(this,"body");this.type=t,this.body=e}static encode(t,e){const i=e?e.length:0,n=4,s=new Uint8Array(n+i);let h=0;return s[h++]=t&255,s[h++]=i>>16&255,s[h++]=i>>8&255,s[h++]=i&255,e&&B.blockCopy(e,0,s,h,i),s}static decode(t){const e=[];for(;t.getLength()-t.getPosition()>=4;){const n=t.readByte(),s=t.readByte()<<16|t.readByte()<<8|t.readByte()>>>0;if(s<0)throw new Error(`type=${n}, length = ${s}, stream=${t.toString()}`);if(t.getLength()<s){t.seek(-4,K.Current);break}const h=new Uint8Array(s);s>0&&t.read(h,0,s);let f=new _(n,h);e.push(f)}return t.tidy(),e}}var A=(o=>(o[o.Handshake=1]="Handshake",o[o.HandshakeAck=2]="HandshakeAck",o[o.Heartbeat=3]="Heartbeat",o[o.Data=4]="Data",o[o.Kick=5]="Kick",o))(A||{});function O(o){const t=new ArrayBuffer(o.length*3),e=new Uint8Array(t);let i=0;for(let s=0;s<o.length;s++){const h=o.charCodeAt(s);let f;h<=127?f=[h]:h<=2047?f=[192|h>>6,128|h&63]:f=[224|h>>12,128|(h&4032)>>6,128|h&63];for(let m=0;m<f.length;m++)e[i]=f[m],++i}const n=new Uint8Array(i);return B.blockCopy(e,0,n,0,i),n}function M(o){const t=new Uint8Array(o),e=[];let i=0,n=0;const s=t.length;for(;i<s;)t[i]<128?(n=t[i],i+=1):t[i]<224?(n=((t[i]&63)<<6)+(t[i+1]&63),i+=2):(n=((t[i]&15)<<12)+((t[i+1]&63)<<6)+(t[i+2]&63),i+=3),e.push(n);return ct(e)}function ct(o){const e=[];for(let i=0;i<o.length;i+=32768)e.push(String.fromCharCode.apply(null,o.slice(i,i+32768)));return e.join("")}var T=(o=>(o[o.Request=0]="Request",o[o.Notify=1]="Notify",o[o.Response=2]="Response",o[o.Push=3]="Push",o[o.Count=4]="Count",o))(T||{});const u=class{constructor(t,e,i,n,s){r(this,"id");r(this,"type");r(this,"compressRoute");r(this,"route");r(this,"body");this.id=t,this.type=e,this.compressRoute=i,this.route=n,this.body=s}static encode(t,e,i,n,s){const h=u.hasId(e)?u.calculateMsgIdBytes(t):0;let f=u.MSG_FLAG_BYTES+h;if(u.hasRoute(e)){if(i){if(typeof n!="number")throw new Error("error flag for number route!");f+=u.MSG_ROUTE_CODE_BYTES}else if(f+=u.MSG_ROUTE_LEN_BYTES,n){if(n=O(n),n.length>255)throw new Error("route maxlength is overflow");f+=n.length}}s&&(f+=s.length);const m=new Uint8Array(f);let p=0;return p=u.encodeMsgFlag(e,i,m,p),u.hasId(e)&&(p=u.encodeMsgId(t,m,p)),u.hasRoute(e)&&(p=u.encodeMsgRoute(i,n,m,p)),s!=null&&(p=u.encodeMsgBody(s,m,p)),m}static decode(t){const e=new Uint8Array(t),i=e.length||e.byteLength;let n=0,s=0,h="";const f=e[n++],m=f&u.MSG_COMPRESS_ROUTE_MASK,p=f>>1&u.MSG_TYPE_MASK;if(u.hasId(p)){let g=e[n],H=0;do g=e[n],s=s+(g&127)*Math.pow(2,7*H),n++,H++;while(g>=128)}if(u.hasRoute(p))if(m)h=(e[n++]<<8|e[n++]).toString();else{const g=e[n++];if(g){let H=new Uint8Array(g);B.blockCopy(e,n,H,0,g),h=M(h)}else h="";n+=g}const E=i-n,w=new Uint8Array(E);return B.blockCopy(e,n,w,0,E),new u(s,p,m,h,w)}static calculateMsgIdBytes(t){let e=0;do e+=1,t>>=7;while(t>0);return e}static encodeMsgFlag(t,e,i,n){if(!u.isValid(t))throw new Error("unknown message type: "+t);return i[n]=t<<1|(e?1:0),n+u.MSG_FLAG_BYTES}static encodeMsgId(t,e,i){do{let n=t%128;const s=Math.floor(t/128);s!==0&&(n=n+128),e[i++]=n,t=s}while(t!==0);return i}static encodeMsgRoute(t,e,i,n){if(t){if(e>u.MSG_ROUTE_CODE_MAX)throw new Error("route number is overflow");i[n++]=e>>8&255,i[n++]=e&255}else e?(i[n++]=e.length&255,B.blockCopy(e,0,i,n,e.length),n+=e.length):i[n++]=0;return n}static encodeMsgBody(t,e,i){return B.blockCopy(t,0,e,i,t.length),i+t.length}static hasId(t){return t===T.Request||t===T.Response}static hasRoute(t){return t===T.Request||t===T.Notify||t===T.Push}static isValid(t){return t>=T.Request&&t<T.Count}};let x=u;r(x,"MSG_FLAG_BYTES",1),r(x,"MSG_ROUTE_CODE_BYTES",2),r(x,"MSG_ID_MAX_BYTES",5),r(x,"MSG_ROUTE_LEN_BYTES",1),r(x,"MSG_ROUTE_CODE_MAX",65535),r(x,"MSG_COMPRESS_ROUTE_MASK",1),r(x,"MSG_TYPE_MASK",7);class dt{constructor(){r(this,"handleHeartBeat",t=>{!this.heartbeatInterval||(this.heartbeatTimeoutId&&(clearTimeout(this.heartbeatTimeoutId),this.heartbeatTimeoutId=null),!this.heartbeatId&&(this.heartbeatId=setTimeout(()=>{this.heartbeatId=null;const e=_.encode(A.Heartbeat);this.send(e),this.nextHeartbeatTimeout=Date.now()+this.heartbeatTimeout,this.heartbeatTimeoutId=setTimeout(this.heartbeatTimeoutCb,this.heartbeatTimeout)},this.heartbeatInterval)))});r(this,"handleHandshake",t=>{let e=JSON.parse(M(t));const i=501;if(e.code===i){this.emit("error","client version not fullfill");return}const n=200;if(e.code!==n){this.emit("error","handshake fail");return}this.handshakeInit(e);const s=_.encode(A.HandshakeAck);this.send(s)});r(this,"handleData",t=>{let e=t;this.decode&&(e=this.decode(e)),this.processMessage(e)});r(this,"handleKick",t=>{t=JSON.parse(M(t)),this.emit("onKick",t)});r(this,"socket",null);r(this,"buffer",new F(8));r(this,"useCrypto",!1);r(this,"encode");r(this,"decode");r(this,"requestIdGenerator",0);r(this,"reconnectUrl","");r(this,"reconnect",!1);r(this,"reconnectTimer");r(this,"reconnectAttempts",0);r(this,"reconnectionDelay",5e3);r(this,"handshakeBuffer",{sys:{type:"js-websocket",version:"0.0.1",rsa:{}},user:{}});r(this,"pushHandlers",new Map);r(this,"handlers",new Map);r(this,"routeMap",new Map);r(this,"callbacks",new Map);r(this,"abbrs",{});r(this,"dict",{});r(this,"heartbeatInterval",0);r(this,"heartbeatTimeout",0);r(this,"nextHeartbeatTimeout",0);r(this,"heartbeatTimeoutId");r(this,"heartbeatId");r(this,"handshakeCallback")}on(t,e){this.pushHandlers[t]=e}emit(t,e=""){const i=this.pushHandlers[t];i!=null&&i(e)}processPackages(t){for(let e=0;e<t.length;e++){const i=t[e],n=this.handlers[i.type];n!=null&&n(i.body)}}defaultDecode(t){const e=x.decode(t);if(!(e.id>0&&(e.route=this.routeMap[e.id],this.routeMap.delete(e.id),!e.route)))return e.body=this.decompose(e),e}decompose(t){let e=t.route;if(t.compressRoute){if(!this.abbrs[e])return{};e=t.route=this.abbrs[e]}return JSON.parse(M(t.body))}reset(){this.reconnect=!1,this.reconnectionDelay=1e3*5,this.reconnectAttempts=0,clearTimeout(this.reconnectTimer)}initData(t){if(!(!t||!t.sys)&&(this.dict=t.sys.dict,this.dict)){this.abbrs={};for(const e in this.dict)this.abbrs[this.dict[e]]=e}}handshakeInit(t){t.sys&&t.sys.heartbeat?(this.heartbeatInterval=t.sys.heartbeat*1e3,this.heartbeatTimeout=this.heartbeatInterval*2):(this.heartbeatInterval=0,this.heartbeatTimeout=0),this.initData(t),typeof this.handshakeCallback=="function"&&this.handshakeCallback(t.user)}processMessage(t){if(t.id){const e=this.callbacks[t.id];this.callbacks.delete(t.id),typeof e=="function"&&e(t.body)}else{const e=this.pushHandlers[t.route];typeof e!="undefined"&&e(t.body)}}heartbeatTimeoutCb(){const t=this.nextHeartbeatTimeout-Date.now();t>100?this.heartbeatTimeoutId=setTimeout(this.heartbeatTimeoutCb,t):(console.error("server heartbeat timeout"),this.emit("heartbeat timeout"),this.disconnect())}send(t){this.socket!=null?this.socket.send(t.buffer):console.log("socket = null")}sendMessage(t,e,i){let n=i;this.encode&&(n=this.encode(t,e,i));const s=_.encode(A.Data,n);this.send(s)}connectInner(t,e,i){console.log("connect to: "+e),t=t||{};const n=10,s=t.maxReconnectAttempts||n;this.reconnectUrl=e;const h=w=>{this.reconnect&&this.emit("reconnect"),this.reset();const S=_.encode(A.Handshake,O(JSON.stringify(this.handshakeBuffer)));this.send(S),i!=null&&i()},f=w=>{let S=new Uint8Array(w.data),g=this.buffer;g.write(S,0,S.length),g.setPosition(0),this.processPackages(_.decode(g)),this.heartbeatTimeout&&(this.nextHeartbeatTimeout=Date.now()+this.heartbeatTimeout)},m=w=>{this.emit("io-error",w),console.error("socket error: ",w)},p=w=>{this.emit("close",w),this.emit("disconnect",w),console.log("socket close: ",w),t.reconnect&&this.reconnectAttempts<s&&(this.reconnect=!0,this.reconnectAttempts++,this.reconnectTimer=setTimeout(()=>{this.connectInner(t,this.reconnectUrl,i)},this.reconnectionDelay),this.reconnectionDelay*=2)};let E=new WebSocket(e);E.binaryType="arraybuffer",E.onopen=h,E.onmessage=f,E.onerror=m,E.onclose=p,this.socket=E}connect(t,e){this.handshakeCallback=t.handshakeCallback,this.encode=t.encode||this.defaultEncode,this.decode=t.decode||this.defaultDecode,this.handshakeBuffer.user=t.user,this.handlers[A.Heartbeat]=this.handleHeartBeat,this.handlers[A.Handshake]=this.handleHandshake,this.handlers[A.Data]=this.handleData,this.handlers[A.Kick]=this.handleKick,this.connectInner(t,t.url,e)}defaultEncode(t,e,i){const n=t!=0?T.Request:T.Notify;i=O(JSON.stringify(i));let s=!1;return this.dict&&this.dict[e]&&(e=this.dict[e],s=!0),x.encode(t,n,s,e,i)}disconnect(){this.socket!=null&&(this.socket.close(),console.log("disconnect"),this.socket=null),this.heartbeatId&&(clearTimeout(this.heartbeatId),this.heartbeatId=null),this.heartbeatTimeoutId&&(clearTimeout(this.heartbeatTimeoutId),this.heartbeatTimeoutId=null)}request(t,e,i){let n=++this.requestIdGenerator;this.sendMessage(n,t,e),this.callbacks[n]=i,this.routeMap[n]=t}notify(t,e){e=e||{},this.sendMessage(0,t,e)}}function ut(o){const t=document.getElementById("mainPanel");if(t==null)return;let e=t.scrollTop>t.scrollHeight-t.clientHeight-1;t.appendChild(o),e&&(t.scrollTop=t.scrollHeight-t.clientHeight)}function D(o){const t=document.createElement("div");t.innerHTML=o,ut(t)}function L(){D("<br>")}function C(o){D("["+Z(new Date).format("HH:mm:ss.S")+"] "+o)}class ft{constructor(){r(this,"currentIndex",-1);r(this,"list",[]);const t="list",e=localStorage.getItem(t);if(e){const i=JSON.parse(e);i&&(this.list=i,this.currentIndex=this.list.length)}window.onunload=i=>{const n=this.list.slice(-100);localStorage.setItem(t,JSON.stringify(n))}}add(t){if(t!=null&&t!=""){const e=this.list,i=e.length;i==0||e[i-1]!==t?this.currentIndex=e.push(t):this.currentIndex=e.length}}getHistories(){return this.list}getHistory(t){return t>=0&&t<this.list.length?this.list[t]:""}move(t){if(t!=0){let e=this.currentIndex+t;if(e>=0&&e<this.list.length)return this.currentIndex=e,this.list[e]}return""}toString(){return`currentIndex=${this.currentIndex}, list=[${this.list}]`}}class bt{constructor(){r(this,"host");r(this,"directory");r(this,"websocketPath");r(this,"autoLoginLimit");r(this,"body");if(document.title!="{{.Data}}"){let t=JSON.parse(document.title);console.log("data:",t),this.host=window.location.host,this.directory=t.directory,this.autoLoginLimit=t.autoLoginLimit,this.websocketPath=t.websocketPath,document.title=t.title,this.body=t.body}else this.host="localhost:8888",this.directory="ws",this.autoLoginLimit=864e5,this.websocketPath="",this.body="<h2>fake body</h2>"}getWebsocketUrl(){return`${document.location.protocol==="https:"?"wss:":"ws:"}//${this.host}/${this.directory}/${this.websocketPath}`}}class pt{constructor(t){r(this,"sendLogin");r(this,"key","autoLoginUser");this.sendLogin=t}login(t,e,i){this.doLogin(t,e),this.save(t,e,i)}tryAutoLogin(){const t=localStorage.getItem(this.key);if(t){const e=JSON.parse(t);e&&new Date().getTime()<e.expireTime&&this.doLogin(e.username,e.password)}}doLogin(t,e){const i="hey pet!",n=Q.exports.sha256.hmac(i,e),s={command:"auth "+t+" "+n};this.sendLogin(s)}save(t,e,i){const n={username:t,password:e,expireTime:new Date().getTime()+i},s=JSON.stringify(n);localStorage.setItem(this.key,s)}}const gt=$("div",{id:"mainPanel"},null,-1),yt={id:"inputBoxDiv"},mt=["onKeydown"],wt=tt({setup(o){let t="",e="",i=!1,n=new bt,s=new ft,h=new dt,f=`${document.location.protocol}//${n.host}/${n.directory}`,m=n.getWebsocketUrl(),p=new pt(a=>{S("console.command",a,g)});h.connect({url:m},()=>{console.log("websocket connected"),D(n.body),L(),p.tryAutoLogin()}),h.on("disconnect",()=>{C("<b> disconnected from server </b>")}),h.on("console.html",E),h.on("console.default",w),window.onload=()=>{const a=document.getElementById("inputBox");!a||(a.focus(),document.onkeydown=function(l){if(l.key==="Enter"&&document.activeElement!==a&&a)return a.focus(),!1})};function E(a){C("<b>server\u54CD\u5E94\uFF1A</b>"+a.data),L()}function w(a){const l=JSON.stringify(a);C("<b>server\u54CD\u5E94\uFF1A</b>"+l),L()}function S(a,l,c){const d=JSON.stringify(l);C("<b>client\u8BF7\u6C42\uFF1A</b>"),D(d),L(),h.request(a,l,c)}function g(a){switch(a.op){case"log.list":j(a.data);break;case"history":q(a.data);break;case"html":E(a);break;case"empty":break;default:w(a)}}function H(a){let l=a.target.value.trim();if(l!==""){if(a.target.value="",l.startsWith("!")){const b=parseInt(l.substring(1))-1;isNaN(b)||(l=s.getHistory(b))}let d=l.split(/\s+/),k=d.length;const y=d[0];if(y==="help"){const b={command:y+" "+f};S("console.command",b,g),s.add(l)}else if(k>=2&&(y==="sub"||y==="unsub")){const b={topic:d[1]},I="console."+y;S(I,b,g),s.add(l)}else if(k>=2&&y==="auth")e=d[1],i=!0,a.target.type="password",C(l+"<br/> <h3>\u8BF7\u8F93\u5165\u5BC6\u7801\uFF1A</h3><br/>"),s.add(l);else if(i&&k>=1)i=!1,a.target.type="text",p.login(e,y,n.autoLoginLimit);else{const b={command:d.join(" ")};S("console.command",b,g),s.add(l)}}else C("");const c=document.getElementById("mainPanel");c&&(c.scrollTop=c.scrollHeight-c.clientHeight)}function q(a){const l=s.getHistories(),c=l.length,d=new Array(c);for(let y=0;y<c;y++)d[y]="<li>"+l[y]+"</li>";let k="<b>\u5386\u53F2\u547D\u4EE4\u5217\u8868\uFF1A</b> <br/> count:&nbsp;"+c+"<br/><ol>"+d.join("")+"</ol>";C(k),L()}function J(a){const l=a.target.value.trim();if(l.length>0){const c={head:l};h.request("console.hint",c,d=>{const k=d.names,y=d.notes,b=k.length;if(b>0&&(a.target.value=z(k),b>1)){const I=new Array(b);for(let v=0;v<b;v++)I[v]=`<tr> <td>${v+1}</td> <td>${k[v]}</td> <td>${y[v]}</td> </tr>`;const W="<table> <tr> <th></th> <th>Name</th> <th>Note</th> </tr>"+I.join("")+"</table>";C(W),L()}})}}function Y(a){a.key;const l=a.key=="ArrowUp"?-1:1,c=s.move(l);if(c!=""){let d=a.target;d.value=c,setTimeout(()=>{let k=c.length;d.setSelectionRange(k,k),d.focus()},0)}}function z(a){if(a.length<2)return a.join();let l=a[0];for(let c=1;c<a.length;c++)for(let d=l.length;d>0&&l!==a[c].substring(0,d);d--)l=l.substring(0,d-1);return l}function j(a){const l=a.logFiles,c=l.length,d=new Array(c);let k=0;for(let b=0;b<c;b++){const I=l[b];k+=I.size;let G=P(I.size);d[b]=`<tr> <td>${b+1}</td> <td>${G}</td> <td> <div class="tips"><a href="${f}/${I.path}">${I.path}</a> <span class="tips_text">${I.sample}</span>
                                <input type="button" class="copy_button" onclick="navigator.clipboard.writeText('${I.path}')" value="\u590D\u5236"/>
                                </div></td> <td>${I.mod_time}</td> </tr>`}let y="<b>\u65E5\u5FD7\u6587\u4EF6\u5217\u8868\uFF1A</b> <br> count:&nbsp;"+c+"<br>total:&nbsp;&nbsp;"+P(k)+"<br>";y+="<table> <tr> <th></th> <th>Size</th> <th>Name</th> <th>Modified Time</th> </tr>"+d.join("")+"</table>",C(y),L()}function P(a){return a<1024?a+"B":a<1048576?(a/1024).toFixed(1)+"K":(a/1048576).toFixed(1)+"M"}return(a,l)=>(at(),et(rt,null,[gt,$("div",yt,[it($("input",{id:"inputBox",ref:"mainPanel","onUpdate:modelValue":l[0]||(l[0]=c=>ot(t)?t.value=c:t=c),placeholder:"Tab\u8865\u5168\u547D\u4EE4, Enter\u6267\u884C\u547D\u4EE4",onKeydown:[U(N(H,["prevent"]),["enter"]),U(N(J,["prevent"]),["tab"]),U(N(Y,["prevent"]),["up","down"])]},null,40,mt),[[nt,st(t)]])])],64))}});ht(wt).mount("#app");
