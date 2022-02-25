var ct=Object.defineProperty;var lt=(s,t,e)=>t in s?ct(s,t,{enumerable:!0,configurable:!0,writable:!0,value:e}):s[t]=e;var r=(s,t,e)=>(lt(s,typeof t!="symbol"?t+"":t,e),e);import{h as j,s as ut,d as dt,u as ft,a as V,o as bt,b as D,c as R,e as X,t as W,f as G,g as $,F as K,r as pt,i as yt,w as gt,v as mt,j as wt,k as F,l as O,m as Z,n as kt}from"./vendor.3ce3a7a4.js";const Et=function(){const t=document.createElement("link").relList;if(t&&t.supports&&t.supports("modulepreload"))return;for(const n of document.querySelectorAll('link[rel="modulepreload"]'))i(n);new MutationObserver(n=>{for(const o of n)if(o.type==="childList")for(const h of o.addedNodes)h.tagName==="LINK"&&h.rel==="modulepreload"&&i(h)}).observe(document,{childList:!0,subtree:!0});function e(n){const o={};return n.integrity&&(o.integrity=n.integrity),n.referrerpolicy&&(o.referrerPolicy=n.referrerpolicy),n.crossorigin==="use-credentials"?o.credentials="include":n.crossorigin==="anonymous"?o.credentials="omit":o.credentials="same-origin",o}function i(n){if(n.ep)return;n.ep=!0;const o=e(n);fetch(n.href,o)}};Et();class C{static blockCopy(t,e,i,n,o){for(let h=0;h<o;h++)i[n++]=t[e++]}}var Q=(s=>(s[s.Begin=0]="Begin",s[s.Current=1]="Current",s[s.End=2]="End",s))(Q||{});const N=class{constructor(t){r(this,"initialIndex",0);r(this,"dirtyBytes",0);r(this,"position",0);r(this,"length",0);r(this,"capacity",0);r(this,"buffer");if(t<0)throw new Error(`capacity=${t}`);this.capacity=t,this.buffer=new Uint8Array(t)}setCapacity(t){if(t!=this.capacity&&t!=this.buffer.length){let e;if(t!=0){e=new Uint8Array(t);for(let i=0;i<this.length;i++)e[i]=this.buffer[i];this.buffer=e}this.dirtyBytes=0,this.capacity=t}}getCapacity(){return this.capacity-this.initialIndex}setPosition(t){if(t<0)throw new Error(`position=${t}`);this.position=this.initialIndex+t}getPosition(){return this.position-this.initialIndex}setLength(t){if(t>this.capacity)throw new Error("length can not be greater than capacity");if(t<0||t+this.initialIndex>N.maxCapacity)throw new Error("out of range");let e=t+this.initialIndex;e>this.length?this.expand(e):e<this.length&&(this.dirtyBytes+=this.length-e),this.length=e,this.position>this.length&&(this.position=this.length)}getLength(){return this.length-this.initialIndex}expand(t){if(t>this.capacity){let e=t;e<32?e=32:e<this.capacity<<1&&(e=this.capacity<<1),this.setCapacity(e)}else if(this.dirtyBytes>0){for(let e=0;e<this.dirtyBytes;e++){let i=e+this.length;this.buffer[i]=0}this.dirtyBytes=0}}readByte(){return this.position>=this.length?-1:this.buffer[this.position++]}read(t,e,i){if(e<0||i<0)throw new Error(`offset=${e}, count=${i}`);if(this.buffer.length-e<i)throw new Error("the size of buffer is less than offset + count");return this.position>=this.length||i==0?0:(this.position>=this.length-i&&(i=this.length-this.position),C.blockCopy(this.buffer,this.position,t,e,i),this.position+=i,i)}write(t,e,i){if(e<0||i<0)throw new Error(`offset=${e}, count=${i}`);if(t.byteLength-e<i)throw new Error("the size of the buffer is less than offset + count");this.position>this.length-i&&this.expand(this.position+i),C.blockCopy(t,e,this.buffer,this.position,i),this.position+=i,this.position>=this.length&&(this.length=this.position)}seek(t,e){if(t>N.maxCapacity)throw new Error("offset is out of range");let i;switch(e){case 0:if(t<0)throw new Error("attempted to seek before start of OctetsSteam");i=this.initialIndex;break;case 1:i=this.position;break;case 2:i=this.length;break;default:throw new Error("invalid SeekOrigin")}if(i+=t,i<this.initialIndex)throw new Error("attempted to seek before start of OctetsStream");return this.position=i,this.position}tidy(){const t=this.length-this.position;C.blockCopy(this.buffer,this.position,this.buffer,0,t),this.setPosition(0),this.setLength(t)}toString(){return`dirtyBytes=${this.dirtyBytes}, position=${this.position}, length=${this.length}, capacity=${this.capacity}, buffer=${this.buffer}`}};let q=N;r(q,"maxCapacity",2147483647);class B{constructor(t,e){r(this,"type");r(this,"body");this.type=t,this.body=e}static encode(t,e){const i=e?e.length:0,n=4,o=new Uint8Array(n+i);let h=0;return o[h++]=t&255,o[h++]=i>>16&255,o[h++]=i>>8&255,o[h++]=i&255,e&&C.blockCopy(e,0,o,h,i),o}static decode(t){const e=[];for(;t.getLength()-t.getPosition()>=4;){const n=t.readByte(),o=t.readByte()<<16|t.readByte()<<8|t.readByte()>>>0;if(o<0)throw new Error(`type=${n}, length = ${o}, stream=${t.toString()}`);if(t.getLength()<o){t.seek(-4,Q.Current);break}const h=new Uint8Array(o);o>0&&t.read(h,0,o);let d=new B(n,h);e.push(d)}return t.tidy(),e}}var I=(s=>(s[s.Handshake=1]="Handshake",s[s.HandshakeAck=2]="HandshakeAck",s[s.Heartbeat=3]="Heartbeat",s[s.Data=4]="Data",s[s.Kick=5]="Kick",s))(I||{});function z(s){const t=new ArrayBuffer(s.length*3),e=new Uint8Array(t);let i=0;for(let o=0;o<s.length;o++){const h=s.charCodeAt(o);let d;h<=127?d=[h]:h<=2047?d=[192|h>>6,128|h&63]:d=[224|h>>12,128|(h&4032)>>6,128|h&63];for(let p=0;p<d.length;p++)e[i]=d[p],++i}const n=new Uint8Array(i);return C.blockCopy(e,0,n,0,i),n}function U(s){const t=new Uint8Array(s),e=[];let i=0,n=0;const o=t.length;for(;i<o;)t[i]<128?(n=t[i],i+=1):t[i]<224?(n=((t[i]&63)<<6)+(t[i+1]&63),i+=2):(n=((t[i]&15)<<12)+((t[i+1]&63)<<6)+(t[i+2]&63),i+=3),e.push(n);return St(e)}function St(s){const e=[];for(let i=0;i<s.length;i+=32768)e.push(String.fromCharCode.apply(null,s.slice(i,i+32768)));return e.join("")}var x=(s=>(s[s.Request=0]="Request",s[s.Notify=1]="Notify",s[s.Response=2]="Response",s[s.Push=3]="Push",s[s.Count=4]="Count",s))(x||{});const u=class{constructor(t,e,i,n,o){r(this,"id");r(this,"type");r(this,"compressRoute");r(this,"route");r(this,"body");this.id=t,this.type=e,this.compressRoute=i,this.route=n,this.body=o}static encode(t,e,i,n,o){const h=u.hasId(e)?u.calculateMsgIdBytes(t):0;let d=u.MSG_FLAG_BYTES+h;if(u.hasRoute(e)){if(i){if(typeof n!="number")throw new Error("error flag for number route!");d+=u.MSG_ROUTE_CODE_BYTES}else if(d+=u.MSG_ROUTE_LEN_BYTES,n){if(n=z(n),n.length>255)throw new Error("route maxlength is overflow");d+=n.length}}o&&(d+=o.length);const p=new Uint8Array(d);let g=0;return g=u.encodeMsgFlag(e,i,p,g),u.hasId(e)&&(g=u.encodeMsgId(t,p,g)),u.hasRoute(e)&&(g=u.encodeMsgRoute(i,n,p,g)),o!=null&&(g=u.encodeMsgBody(o,p,g)),p}static decode(t){const e=new Uint8Array(t),i=e.length||e.byteLength;let n=0,o=0,h="";const d=e[n++],p=d&u.MSG_COMPRESS_ROUTE_MASK,g=d>>1&u.MSG_TYPE_MASK;if(u.hasId(g)){let y=e[n],A=0;do y=e[n],o=o+(y&127)*Math.pow(2,7*A),n++,A++;while(y>=128)}if(u.hasRoute(g))if(p!=0)h=(e[n++]<<8|e[n++]).toString();else{const y=e[n++];if(y>0){let A=new Uint8Array(y);C.blockCopy(e,n,A,0,y),h=U(A)}else h="";n+=y}const w=i-n,m=new Uint8Array(w);return C.blockCopy(e,n,m,0,w),new u(o,g,p,h,m)}static calculateMsgIdBytes(t){let e=0;do e+=1,t>>=7;while(t>0);return e}static encodeMsgFlag(t,e,i,n){if(!u.isValid(t))throw new Error("unknown message type: "+t);return i[n]=t<<1|(e?1:0),n+u.MSG_FLAG_BYTES}static encodeMsgId(t,e,i){do{let n=t%128;const o=Math.floor(t/128);o!==0&&(n=n+128),e[i++]=n,t=o}while(t!==0);return i}static encodeMsgRoute(t,e,i,n){if(t){if(e>u.MSG_ROUTE_CODE_MAX)throw new Error("route number is overflow");i[n++]=e>>8&255,i[n++]=e&255}else e?(i[n++]=e.length&255,C.blockCopy(e,0,i,n,e.length),n+=e.length):i[n++]=0;return n}static encodeMsgBody(t,e,i){return C.blockCopy(t,0,e,i,t.length),i+t.length}static hasId(t){return t===x.Request||t===x.Response}static hasRoute(t){return t===x.Request||t===x.Notify||t===x.Push}static isValid(t){return t>=x.Request&&t<x.Count}};let _=u;r(_,"MSG_FLAG_BYTES",1),r(_,"MSG_ROUTE_CODE_BYTES",2),r(_,"MSG_ID_MAX_BYTES",5),r(_,"MSG_ROUTE_LEN_BYTES",1),r(_,"MSG_ROUTE_CODE_MAX",65535),r(_,"MSG_COMPRESS_ROUTE_MASK",1),r(_,"MSG_TYPE_MASK",7);class Tt{constructor(){r(this,"handleHeartBeat",t=>{!this.heartbeatInterval||(this.heartbeatTimeoutId&&(clearTimeout(this.heartbeatTimeoutId),this.heartbeatTimeoutId=null),!this.heartbeatId&&(this.heartbeatId=setTimeout(()=>{this.heartbeatId=null;const e=B.encode(I.Heartbeat);this.send(e),this.nextHeartbeatTimeout=Date.now()+this.heartbeatTimeout,this.heartbeatTimeoutId=setTimeout(this.heartbeatTimeoutCallback,this.heartbeatTimeout)},this.heartbeatInterval)))});r(this,"handleHandshake",t=>{let e=JSON.parse(U(t));const i=501;if(e.code===i){this.emit("error","client version not fullfill");return}const n=200;if(e.code!==n){this.emit("error","handshake fail");return}this.handshakeInit(e);const o=B.encode(I.HandshakeAck);this.send(o)});r(this,"handleData",t=>{let e=t;this.decode&&(e=this.decode(e)),this.processMessage(e)});r(this,"handleKick",t=>{t=JSON.parse(U(t)),this.emit("onKick",t)});r(this,"socket",null);r(this,"buffer",new q(8));r(this,"useCrypto",!1);r(this,"encode");r(this,"decode");r(this,"requestIdGenerator",0);r(this,"reconnectUrl","");r(this,"reconnect",!1);r(this,"reconnectTimer");r(this,"reconnectAttempts",0);r(this,"reconnectionDelay",5e3);r(this,"handshakeBuffer",{sys:{type:"js-websocket",version:"0.0.1",rsa:{}},user:{}});r(this,"pushHandlers",new Map);r(this,"handlers",new Map);r(this,"routeMap",new Map);r(this,"callbacks",new Map);r(this,"abbrs",{});r(this,"dict",{});r(this,"heartbeatInterval",0);r(this,"heartbeatTimeout",0);r(this,"nextHeartbeatTimeout",0);r(this,"heartbeatTimeoutId");r(this,"heartbeatId");r(this,"handshakeCallback")}on(t,e){this.pushHandlers[t]=e}emit(t,e=""){const i=this.pushHandlers[t];i!=null&&i(e)}processPackages(t){for(let e=0;e<t.length;e++){const i=t[e],n=this.handlers[i.type];n!=null&&n(i.body)}}defaultDecode(t){const e=_.decode(t);if(!(e.id>0&&(e.route=this.routeMap[e.id],this.routeMap.delete(e.id),!e.route)))return e.body=this.decompose(e),e}decompose(t){let e=t.route;if(t.compressRoute){if(!this.abbrs[e])return{};e=t.route=this.abbrs[e]}return JSON.parse(U(t.body))}reset(){this.reconnect=!1,this.reconnectionDelay=1e3*5,this.reconnectAttempts=0,clearTimeout(this.reconnectTimer)}initData(t){if(!(!t||!t.sys)&&(this.dict=t.sys.dict,this.dict)){this.abbrs={};for(const e in this.dict)this.abbrs[this.dict[e]]=e}}handshakeInit(t){t.sys&&t.sys.heartbeat?(this.heartbeatInterval=t.sys.heartbeat*1e3,this.heartbeatTimeout=this.heartbeatInterval*2):(this.heartbeatInterval=0,this.heartbeatTimeout=0),this.initData(t),typeof this.handshakeCallback=="function"&&this.handshakeCallback(t.user)}processMessage(t){if(t.id){const e=this.callbacks[t.id];this.callbacks.delete(t.id),typeof e=="function"&&e(t.body)}else{const e=this.pushHandlers[t.route];typeof e!="undefined"?e(t.body):console.log(`cannot find handler for route=${t.route}, msg=`,t)}}heartbeatTimeoutCallback(){const t=this.nextHeartbeatTimeout-Date.now();t>100?this.heartbeatTimeoutId=setTimeout(this.heartbeatTimeoutCallback,t):(console.error("server heartbeat timeout"),this.emit("heartbeat timeout"),this.disconnect())}send(t){this.socket!=null?this.socket.send(t.buffer):console.log("socket = null")}sendMessage(t,e,i){let n=i;this.encode&&(n=this.encode(t,e,i));const o=B.encode(I.Data,n);this.send(o)}connectInner(t,e,i){console.log("connect to: "+e),t=t||{};const n=10,o=t.maxReconnectAttempts||n;this.reconnectUrl=e;const h=m=>{this.reconnect&&this.emit("reconnect"),this.reset();const S=B.encode(I.Handshake,z(JSON.stringify(this.handshakeBuffer)));this.send(S),i!=null&&i()},d=m=>{let S=new Uint8Array(m.data),y=this.buffer;y.write(S,0,S.length),y.setPosition(0),this.processPackages(B.decode(y)),this.heartbeatTimeout&&(this.nextHeartbeatTimeout=Date.now()+this.heartbeatTimeout)},p=m=>{this.emit("io-error",m),console.error("socket error: ",m)},g=m=>{this.emit("close",m),this.emit("disconnect",m),console.log("socket close: ",m),t.reconnect&&this.reconnectAttempts<o&&(this.reconnect=!0,this.reconnectAttempts++,this.reconnectTimer=setTimeout(()=>{this.connectInner(t,this.reconnectUrl,i)},this.reconnectionDelay),this.reconnectionDelay*=2)};let w=new WebSocket(e);w.binaryType="arraybuffer",w.onopen=h,w.onmessage=d,w.onerror=p,w.onclose=g,this.socket=w}connect(t,e){this.handshakeCallback=t.handshakeCallback,this.encode=t.encode||this.defaultEncode,this.decode=t.decode||this.defaultDecode,this.handshakeBuffer.user=t.user,this.handlers[I.Heartbeat]=this.handleHeartBeat,this.handlers[I.Handshake]=this.handleHandshake,this.handlers[I.Data]=this.handleData,this.handlers[I.Kick]=this.handleKick,this.connectInner(t,t.url,e)}defaultEncode(t,e,i){const n=t!=0?x.Request:x.Notify;i=z(JSON.stringify(i));let o=!1;return this.dict&&this.dict[e]&&(e=this.dict[e],o=!0),_.encode(t,n,o,e,i)}disconnect(){this.socket!=null&&(this.socket.close(),console.log("disconnect"),this.socket=null),this.heartbeatId&&(clearTimeout(this.heartbeatId),this.heartbeatId=null),this.heartbeatTimeoutId&&(clearTimeout(this.heartbeatTimeoutId),this.heartbeatTimeoutId=null)}request(t,e,i){let n=++this.requestIdGenerator;this.sendMessage(n,t,e),this.callbacks[n]=i,this.routeMap[n]=t}notify(t,e){e=e||{},this.sendMessage(0,t,e)}}let P=null;function tt(){return P==null&&(P=document.getElementById("mainPanel")),P}function _t(s){const t=tt();t&&(t.appendChild(s),et())}function et(){const s=tt();if(s){const t=s.scrollHeight-s.clientHeight-1;s.scrollTop<t&&(s.scrollTop=t)}}function v(s){const t=document.createElement("div");return t.innerHTML=s,_t(t),t}function M(){v("<br>")}function L(s){v("["+j(new Date).format("HH:mm:ss.S")+"] "+s)}class xt{constructor(){r(this,"host");r(this,"directory");r(this,"websocketPath");r(this,"autoLoginLimit");r(this,"body");if(document.title!="{{.Data}}"){let t=JSON.parse(document.title);this.host=window.location.host,this.directory=t.directory,this.autoLoginLimit=t.autoLoginLimit,this.websocketPath=t.websocketPath,document.title=t.title,this.body=t.body}else this.host="localhost:8888",this.directory="ws",this.autoLoginLimit=864e5,this.websocketPath="",this.body="<h2>fake body</h2>"}getRootUrl(){let t=`${document.location.protocol}//${this.host}/${this.directory}`;return t.endsWith("/")&&(t=t.substring(0,t.length-1)),t}getWebsocketUrl(){const e=document.location.protocol==="https:"?"wss:":"ws:";return this.directory!=""?`${e}//${this.host}/${this.directory}/${this.websocketPath}`:`${e}//${this.host}/${this.websocketPath}`}toString(){return`host=${this.host}, directory=${this.directory}, websocketPath=${this.websocketPath}, autoLoginLimit=${this.autoLoginLimit}, body=${this.body}`}}class It{constructor(t){r(this,"sendLogin");r(this,"key","autoLoginUser");this.sendLogin=t}login(t,e,i){this.doLogin(t,e),this.save(t,e,i)}tryAutoLogin(){const t=localStorage.getItem(this.key);if(t){const e=JSON.parse(t);e&&new Date().getTime()<e.expireTime&&this.doLogin(e.username,e.password)}}doLogin(t,e){const i="hey pet!",n=ut.exports.sha256.hmac(i,e),o={command:"auth "+t+" "+n};this.sendLogin(o)}save(t,e,i){const n={username:t,password:e,expireTime:new Date().getTime()+i},o=JSON.stringify(n);localStorage.setItem(this.key,o)}}const it=dt({id:"historyStore",state:()=>ft("this.is.history.store",{currentIndex:0,list:[]}),getters:{histories:s=>s.list,count:s=>s.list.length},actions:{add(s){if(s!=null&&s!=""){const t=this.list,e=t.length;e==0||t[e-1]!==s?this.currentIndex=t.push(s):this.currentIndex=t.length}},getHistory(s){return s>=0&&s<this.list.length?this.list[s]:""},move(s){if(s!=0){let t=this.currentIndex+s;if(t>=0&&t<this.list.length)return this.currentIndex=t,this.list[t]}return""}}}),Ct=$("b",null,"\u5386\u53F2\u547D\u4EE4\u5217\u8868\uFF1A",-1),Lt=X(),At=$("br",null,null,-1),Bt={id:"history-with-index"},$t=$("br",null,null,-1),Ht=V({setup(s){const t=it();return bt(()=>{et()}),(e,i)=>(D(),R(K,null,[Ct,Lt,At,X(" count: \xA0 "+W(G(t).count)+" ",1),$("ol",Bt,[(D(!0),R(K,null,pt(G(t).histories,n=>(D(),R("li",{key:n},W(n),1))),128))]),$t],64))}});const Mt=$("div",{id:"mainPanel"},null,-1),vt={id:"inputBoxDiv"},Dt=["onKeydown"],Rt=V({setup(s){let t=yt(""),e="",i=!1,n=new xt;const o=it();let h=new Tt,d=n.getRootUrl(),p=new It(a=>{S("console.command",a,y)});h.connect({url:n.getWebsocketUrl()},()=>{console.log("websocket connected"),v(n.body),M(),p.tryAutoLogin()});const g=new Date;h.on("disconnect",()=>{const a=j.duration(new Date().getTime()-g.getTime(),"milliseconds").humanize();L(`<b> disconnected from server after ${a} </b>`)}),h.on("console.html",w),h.on("console.default",m),window.onload=()=>{const a=document.getElementById("inputBox");!a||(a.focus(),document.onkeydown=function(c){if(c.key==="Enter"&&document.activeElement!==a&&a)return a.focus(),!1})};function w(a){L("<b>server\u54CD\u5E94\uFF1A</b>"+a),M()}function m(a){const c=JSON.stringify(a);L("<b>server\u54CD\u5E94\uFF1A</b>"+c),M()}function S(a,c,l){const f=JSON.stringify(c);L("<b>client\u8BF7\u6C42\uFF1A</b>"),v(f),M(),h.request(a,c,l)}function y(a){switch(a.op){case"log.list":at(a.data);break;case"history":nt(a.data);break;case"html":w(a.data);break;case"empty":break;default:m(a)}}function A(a){let c=t.value;if(c!==""){if(t.value="",c.startsWith("!")){const b=parseInt(c.substring(1))-1;console.log("index:",b),isNaN(b)||(c=o.getHistory(b),c=o.getHistory(b))}let f=c.split(/\s+/),T=f.length;const k=f[0];if(k==="help"){const b={command:k+" "+d};S("console.command",b,y),o.add(c)}else if(T>=2&&(k==="sub"||k==="unsub")){const b={topic:f[1]},E="console."+k;S(E,b,y),o.add(c)}else if(T>=2&&k==="auth")e=f[1],i=!0,a.target.type="password",L(c+"<br/> <h3>\u8BF7\u8F93\u5165\u5BC6\u7801\uFF1A</h3><br/>"),o.add(c);else if(i&&T>=1)i=!1,a.target.type="text",p.login(e,k,n.autoLoginLimit);else{const b={command:f.join(" ")};S("console.command",b,y),o.add(c)}}else L("");const l=document.getElementById("mainPanel");l&&(l.scrollTop=l.scrollHeight-l.clientHeight)}function nt(a){Z(Ht).mount(v(""))}function st(a){const c=t.value;if(c.length>0){const l={head:c};h.request("console.hint",l,f=>{const T=f.names,k=f.notes,b=T.length;if(b>0&&(t.value=rt(T),b>1)){const E=new Array(b);for(let H=0;H<b;H++)E[H]=`<tr> <td>${H+1}</td> <td>${T[H]}</td> <td>${k[H]}</td> </tr>`;const ht="<table> <tr> <th></th> <th>Name</th> <th>Note</th> </tr>"+E.join("")+"</table>";L(ht),M()}})}}function ot(a){const c=a.key=="ArrowUp"?-1:1,l=o.move(c);l!=""&&(t.value=l,setTimeout(()=>{let f=l.length;a.target.setSelectionRange(f,f),a.target.focus()},0))}function rt(a){if(a.length<2)return a.join();let c=a[0];for(let l=1;l<a.length;l++)for(let f=c.length;f>0&&c!==a[l].substring(0,f);f--)c=c.substring(0,f-1);return c}function at(a){const c=a.logFiles,l=c.length,f=new Array(l);let T=0;for(let b=0;b<l;b++){const E=c[b];T+=E.size;let Y=J(E.size);f[b]=`<tr> <td>${b+1}</td> <td>${Y}</td> <td> <div class="tips"><a href="${d}/${E.path}">${E.path}</a> <span class="tips_text">${E.sample}</span>
                                <input type="button" class="copy_button" onclick="copyToClipboard('${E.path}')" value="\u590D\u5236"/>
                                </div></td> <td>${E.mod_time}</td> </tr>`}let k="<b>\u65E5\u5FD7\u6587\u4EF6\u5217\u8868\uFF1A</b> <br> count:&nbsp;"+l+"<br>total:&nbsp;&nbsp;"+J(T)+"<br>";k+="<table> <tr> <th></th> <th>Size</th> <th>Name</th> <th>Modified Time</th> </tr>"+f.join("")+"</table>",L(k),M()}function J(a){return a<1024?a+"B":a<1048576?(a/1024).toFixed(1)+"K":(a/1048576).toFixed(1)+"M"}return(a,c)=>(D(),R(K,null,[Mt,$("div",vt,[gt($("input",{id:"inputBox","onUpdate:modelValue":c[0]||(c[0]=l=>wt(t)?t.value=l:t=l),placeholder:"Tab\u8865\u5168\u547D\u4EE4, Enter\u6267\u884C\u547D\u4EE4",onKeydown:[F(O(A,["prevent"]),["enter"]),F(O(st,["prevent"]),["tab"]),F(O(ot,["prevent"]),["up","down"])]},null,40,Dt),[[mt,G(t)]])])],64))}});Z(Rt).use(kt()).mount("#app");
