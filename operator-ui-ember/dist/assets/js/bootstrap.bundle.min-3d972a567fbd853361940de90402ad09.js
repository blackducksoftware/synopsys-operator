!function(t,e){"object"==typeof exports&&"undefined"!=typeof module?e(exports,require("jquery")):"function"==typeof define&&define.amd?define(["exports","jquery"],e):e((t=t||self).bootstrap={},t.jQuery)}(this,function(t,e){"use strict"
function n(t,e){for(var n=0;n<e.length;n++){var i=e[n]
i.enumerable=i.enumerable||!1,i.configurable=!0,"value"in i&&(i.writable=!0),Object.defineProperty(t,i.key,i)}}function i(t,e,i){return e&&n(t.prototype,e),i&&n(t,i),t}function o(t){for(var e=1;e<arguments.length;e++){var n=null!=arguments[e]?arguments[e]:{},i=Object.keys(n)
"function"==typeof Object.getOwnPropertySymbols&&(i=i.concat(Object.getOwnPropertySymbols(n).filter(function(t){return Object.getOwnPropertyDescriptor(n,t).enumerable}))),i.forEach(function(e){var i,o,r
i=t,r=n[o=e],o in i?Object.defineProperty(i,o,{value:r,enumerable:!0,configurable:!0,writable:!0}):i[o]=r})}return t}e=e&&e.hasOwnProperty("default")?e.default:e
var r="transitionend"
var s={TRANSITION_END:"bsTransitionEnd",getUID:function(t){for(;t+=~~(1e6*Math.random()),document.getElementById(t););return t},getSelectorFromElement:function(t){var e=t.getAttribute("data-target")
if(!e||"#"===e){var n=t.getAttribute("href")
e=n&&"#"!==n?n.trim():""}try{return document.querySelector(e)?e:null}catch(t){return null}},getTransitionDurationFromElement:function(t){if(!t)return 0
var n=e(t).css("transition-duration"),i=e(t).css("transition-delay"),o=parseFloat(n),r=parseFloat(i)
return o||r?(n=n.split(",")[0],i=i.split(",")[0],1e3*(parseFloat(n)+parseFloat(i))):0},reflow:function(t){return t.offsetHeight},triggerTransitionEnd:function(t){e(t).trigger(r)},supportsTransitionEnd:function(){return Boolean(r)},isElement:function(t){return(t[0]||t).nodeType},typeCheckConfig:function(t,e,n){for(var i in n)if(Object.prototype.hasOwnProperty.call(n,i)){var o=n[i],r=e[i],a=r&&s.isElement(r)?"element":(l=r,{}.toString.call(l).match(/\s([a-z]+)/i)[1].toLowerCase())
if(!new RegExp(o).test(a))throw new Error(t.toUpperCase()+': Option "'+i+'" provided type "'+a+'" but expected type "'+o+'".')}var l},findShadowRoot:function(t){if(!document.documentElement.attachShadow)return null
if("function"!=typeof t.getRootNode)return t instanceof ShadowRoot?t:t.parentNode?s.findShadowRoot(t.parentNode):null
var e=t.getRootNode()
return e instanceof ShadowRoot?e:null}}
e.fn.emulateTransitionEnd=function(t){var n=this,i=!1
return e(this).one(s.TRANSITION_END,function(){i=!0}),setTimeout(function(){i||s.triggerTransitionEnd(n)},t),this},e.event.special[s.TRANSITION_END]={bindType:r,delegateType:r,handle:function(t){if(e(t.target).is(this))return t.handleObj.handler.apply(this,arguments)}}
var a="alert",l="bs.alert",c="."+l,h=e.fn[a],u={CLOSE:"close"+c,CLOSED:"closed"+c,CLICK_DATA_API:"click"+c+".data-api"},f=function(){function t(t){this._element=t}var n=t.prototype
return n.close=function(t){var e=this._element
t&&(e=this._getRootElement(t)),this._triggerCloseEvent(e).isDefaultPrevented()||this._removeElement(e)},n.dispose=function(){e.removeData(this._element,l),this._element=null},n._getRootElement=function(t){var n=s.getSelectorFromElement(t),i=!1
return n&&(i=document.querySelector(n)),i||(i=e(t).closest(".alert")[0]),i},n._triggerCloseEvent=function(t){var n=e.Event(u.CLOSE)
return e(t).trigger(n),n},n._removeElement=function(t){var n=this
if(e(t).removeClass("show"),e(t).hasClass("fade")){var i=s.getTransitionDurationFromElement(t)
e(t).one(s.TRANSITION_END,function(e){return n._destroyElement(t,e)}).emulateTransitionEnd(i)}else this._destroyElement(t)},n._destroyElement=function(t){e(t).detach().trigger(u.CLOSED).remove()},t._jQueryInterface=function(n){return this.each(function(){var i=e(this),o=i.data(l)
o||(o=new t(this),i.data(l,o)),"close"===n&&o[n](this)})},t._handleDismiss=function(t){return function(e){e&&e.preventDefault(),t.close(this)}},i(t,null,[{key:"VERSION",get:function(){return"4.3.1"}}]),t}()
e(document).on(u.CLICK_DATA_API,'[data-dismiss="alert"]',f._handleDismiss(new f)),e.fn[a]=f._jQueryInterface,e.fn[a].Constructor=f,e.fn[a].noConflict=function(){return e.fn[a]=h,f._jQueryInterface}
var d="button",p="bs.button",m="."+p,g=".data-api",_=e.fn[d],v="active",y='[data-toggle^="button"]',E=".btn",b={CLICK_DATA_API:"click"+m+g,FOCUS_BLUR_DATA_API:"focus"+m+g+" blur"+m+g},w=function(){function t(t){this._element=t}var n=t.prototype
return n.toggle=function(){var t=!0,n=!0,i=e(this._element).closest('[data-toggle="buttons"]')[0]
if(i){var o=this._element.querySelector('input:not([type="hidden"])')
if(o){if("radio"===o.type)if(o.checked&&this._element.classList.contains(v))t=!1
else{var r=i.querySelector(".active")
r&&e(r).removeClass(v)}if(t){if(o.hasAttribute("disabled")||i.hasAttribute("disabled")||o.classList.contains("disabled")||i.classList.contains("disabled"))return
o.checked=!this._element.classList.contains(v),e(o).trigger("change")}o.focus(),n=!1}}n&&this._element.setAttribute("aria-pressed",!this._element.classList.contains(v)),t&&e(this._element).toggleClass(v)},n.dispose=function(){e.removeData(this._element,p),this._element=null},t._jQueryInterface=function(n){return this.each(function(){var i=e(this).data(p)
i||(i=new t(this),e(this).data(p,i)),"toggle"===n&&i[n]()})},i(t,null,[{key:"VERSION",get:function(){return"4.3.1"}}]),t}()
e(document).on(b.CLICK_DATA_API,y,function(t){t.preventDefault()
var n=t.target
e(n).hasClass("btn")||(n=e(n).closest(E)),w._jQueryInterface.call(e(n),"toggle")}).on(b.FOCUS_BLUR_DATA_API,y,function(t){var n=e(t.target).closest(E)[0]
e(n).toggleClass("focus",/^focus(in)?$/.test(t.type))}),e.fn[d]=w._jQueryInterface,e.fn[d].Constructor=w,e.fn[d].noConflict=function(){return e.fn[d]=_,w._jQueryInterface}
var C="carousel",T="bs.carousel",S="."+T,D=".data-api",I=e.fn[C],A={interval:5e3,keyboard:!0,slide:!1,pause:"hover",wrap:!0,touch:!0},O={interval:"(number|boolean)",keyboard:"boolean",slide:"(boolean|string)",pause:"(string|boolean)",wrap:"boolean",touch:"boolean"},N="next",k="prev",L={SLIDE:"slide"+S,SLID:"slid"+S,KEYDOWN:"keydown"+S,MOUSEENTER:"mouseenter"+S,MOUSELEAVE:"mouseleave"+S,TOUCHSTART:"touchstart"+S,TOUCHMOVE:"touchmove"+S,TOUCHEND:"touchend"+S,POINTERDOWN:"pointerdown"+S,POINTERUP:"pointerup"+S,DRAG_START:"dragstart"+S,LOAD_DATA_API:"load"+S+D,CLICK_DATA_API:"click"+S+D},x="active",P=".active.carousel-item",H=".carousel-indicators",j={TOUCH:"touch",PEN:"pen"},R=function(){function t(t,e){this._items=null,this._interval=null,this._activeElement=null,this._isPaused=!1,this._isSliding=!1,this.touchTimeout=null,this.touchStartX=0,this.touchDeltaX=0,this._config=this._getConfig(e),this._element=t,this._indicatorsElement=this._element.querySelector(H),this._touchSupported="ontouchstart"in document.documentElement||0<navigator.maxTouchPoints,this._pointerEvent=Boolean(window.PointerEvent||window.MSPointerEvent),this._addEventListeners()}var n=t.prototype
return n.next=function(){this._isSliding||this._slide(N)},n.nextWhenVisible=function(){!document.hidden&&e(this._element).is(":visible")&&"hidden"!==e(this._element).css("visibility")&&this.next()},n.prev=function(){this._isSliding||this._slide(k)},n.pause=function(t){t||(this._isPaused=!0),this._element.querySelector(".carousel-item-next, .carousel-item-prev")&&(s.triggerTransitionEnd(this._element),this.cycle(!0)),clearInterval(this._interval),this._interval=null},n.cycle=function(t){t||(this._isPaused=!1),this._interval&&(clearInterval(this._interval),this._interval=null),this._config.interval&&!this._isPaused&&(this._interval=setInterval((document.visibilityState?this.nextWhenVisible:this.next).bind(this),this._config.interval))},n.to=function(t){var n=this
this._activeElement=this._element.querySelector(P)
var i=this._getItemIndex(this._activeElement)
if(!(t>this._items.length-1||t<0))if(this._isSliding)e(this._element).one(L.SLID,function(){return n.to(t)})
else{if(i===t)return this.pause(),void this.cycle()
var o=i<t?N:k
this._slide(o,this._items[t])}},n.dispose=function(){e(this._element).off(S),e.removeData(this._element,T),this._items=null,this._config=null,this._element=null,this._interval=null,this._isPaused=null,this._isSliding=null,this._activeElement=null,this._indicatorsElement=null},n._getConfig=function(t){return t=o({},A,t),s.typeCheckConfig(C,t,O),t},n._handleSwipe=function(){var t=Math.abs(this.touchDeltaX)
if(!(t<=40)){var e=t/this.touchDeltaX
0<e&&this.prev(),e<0&&this.next()}},n._addEventListeners=function(){var t=this
this._config.keyboard&&e(this._element).on(L.KEYDOWN,function(e){return t._keydown(e)}),"hover"===this._config.pause&&e(this._element).on(L.MOUSEENTER,function(e){return t.pause(e)}).on(L.MOUSELEAVE,function(e){return t.cycle(e)}),this._config.touch&&this._addTouchEventListeners()},n._addTouchEventListeners=function(){var t=this
if(this._touchSupported){var n=function(e){t._pointerEvent&&j[e.originalEvent.pointerType.toUpperCase()]?t.touchStartX=e.originalEvent.clientX:t._pointerEvent||(t.touchStartX=e.originalEvent.touches[0].clientX)},i=function(e){t._pointerEvent&&j[e.originalEvent.pointerType.toUpperCase()]&&(t.touchDeltaX=e.originalEvent.clientX-t.touchStartX),t._handleSwipe(),"hover"===t._config.pause&&(t.pause(),t.touchTimeout&&clearTimeout(t.touchTimeout),t.touchTimeout=setTimeout(function(e){return t.cycle(e)},500+t._config.interval))}
e(this._element.querySelectorAll(".carousel-item img")).on(L.DRAG_START,function(t){return t.preventDefault()}),this._pointerEvent?(e(this._element).on(L.POINTERDOWN,function(t){return n(t)}),e(this._element).on(L.POINTERUP,function(t){return i(t)}),this._element.classList.add("pointer-event")):(e(this._element).on(L.TOUCHSTART,function(t){return n(t)}),e(this._element).on(L.TOUCHMOVE,function(e){var n;(n=e).originalEvent.touches&&1<n.originalEvent.touches.length?t.touchDeltaX=0:t.touchDeltaX=n.originalEvent.touches[0].clientX-t.touchStartX}),e(this._element).on(L.TOUCHEND,function(t){return i(t)}))}},n._keydown=function(t){if(!/input|textarea/i.test(t.target.tagName))switch(t.which){case 37:t.preventDefault(),this.prev()
break
case 39:t.preventDefault(),this.next()}},n._getItemIndex=function(t){return this._items=t&&t.parentNode?[].slice.call(t.parentNode.querySelectorAll(".carousel-item")):[],this._items.indexOf(t)},n._getItemByDirection=function(t,e){var n=t===N,i=t===k,o=this._getItemIndex(e),r=this._items.length-1
if((i&&0===o||n&&o===r)&&!this._config.wrap)return e
var s=(o+(t===k?-1:1))%this._items.length
return-1===s?this._items[this._items.length-1]:this._items[s]},n._triggerSlideEvent=function(t,n){var i=this._getItemIndex(t),o=this._getItemIndex(this._element.querySelector(P)),r=e.Event(L.SLIDE,{relatedTarget:t,direction:n,from:o,to:i})
return e(this._element).trigger(r),r},n._setActiveIndicatorElement=function(t){if(this._indicatorsElement){var n=[].slice.call(this._indicatorsElement.querySelectorAll(".active"))
e(n).removeClass(x)
var i=this._indicatorsElement.children[this._getItemIndex(t)]
i&&e(i).addClass(x)}},n._slide=function(t,n){var i,o,r,a=this,l=this._element.querySelector(P),c=this._getItemIndex(l),h=n||l&&this._getItemByDirection(t,l),u=this._getItemIndex(h),f=Boolean(this._interval)
if(r=t===N?(i="carousel-item-left",o="carousel-item-next","left"):(i="carousel-item-right",o="carousel-item-prev","right"),h&&e(h).hasClass(x))this._isSliding=!1
else if(!this._triggerSlideEvent(h,r).isDefaultPrevented()&&l&&h){this._isSliding=!0,f&&this.pause(),this._setActiveIndicatorElement(h)
var d=e.Event(L.SLID,{relatedTarget:h,direction:r,from:c,to:u})
if(e(this._element).hasClass("slide")){e(h).addClass(o),s.reflow(h),e(l).addClass(i),e(h).addClass(i)
var p=parseInt(h.getAttribute("data-interval"),10)
this._config.interval=p?(this._config.defaultInterval=this._config.defaultInterval||this._config.interval,p):this._config.defaultInterval||this._config.interval
var m=s.getTransitionDurationFromElement(l)
e(l).one(s.TRANSITION_END,function(){e(h).removeClass(i+" "+o).addClass(x),e(l).removeClass(x+" "+o+" "+i),a._isSliding=!1,setTimeout(function(){return e(a._element).trigger(d)},0)}).emulateTransitionEnd(m)}else e(l).removeClass(x),e(h).addClass(x),this._isSliding=!1,e(this._element).trigger(d)
f&&this.cycle()}},t._jQueryInterface=function(n){return this.each(function(){var i=e(this).data(T),r=o({},A,e(this).data())
"object"==typeof n&&(r=o({},r,n))
var s="string"==typeof n?n:r.slide
if(i||(i=new t(this,r),e(this).data(T,i)),"number"==typeof n)i.to(n)
else if("string"==typeof s){if(void 0===i[s])throw new TypeError('No method named "'+s+'"')
i[s]()}else r.interval&&r.ride&&(i.pause(),i.cycle())})},t._dataApiClickHandler=function(n){var i=s.getSelectorFromElement(this)
if(i){var r=e(i)[0]
if(r&&e(r).hasClass("carousel")){var a=o({},e(r).data(),e(this).data()),l=this.getAttribute("data-slide-to")
l&&(a.interval=!1),t._jQueryInterface.call(e(r),a),l&&e(r).data(T).to(l),n.preventDefault()}}},i(t,null,[{key:"VERSION",get:function(){return"4.3.1"}},{key:"Default",get:function(){return A}}]),t}()
e(document).on(L.CLICK_DATA_API,"[data-slide], [data-slide-to]",R._dataApiClickHandler),e(window).on(L.LOAD_DATA_API,function(){for(var t=[].slice.call(document.querySelectorAll('[data-ride="carousel"]')),n=0,i=t.length;n<i;n++){var o=e(t[n])
R._jQueryInterface.call(o,o.data())}}),e.fn[C]=R._jQueryInterface,e.fn[C].Constructor=R,e.fn[C].noConflict=function(){return e.fn[C]=I,R._jQueryInterface}
var F="collapse",M="bs.collapse",W="."+M,U=e.fn[F],B={toggle:!0,parent:""},q={toggle:"boolean",parent:"(string|element)"},K={SHOW:"show"+W,SHOWN:"shown"+W,HIDE:"hide"+W,HIDDEN:"hidden"+W,CLICK_DATA_API:"click"+W+".data-api"},Q="show",V="collapse",Y="collapsing",z="collapsed",X='[data-toggle="collapse"]',G=function(){function t(t,e){this._isTransitioning=!1,this._element=t,this._config=this._getConfig(e),this._triggerArray=[].slice.call(document.querySelectorAll('[data-toggle="collapse"][href="#'+t.id+'"],[data-toggle="collapse"][data-target="#'+t.id+'"]'))
for(var n=[].slice.call(document.querySelectorAll(X)),i=0,o=n.length;i<o;i++){var r=n[i],a=s.getSelectorFromElement(r),l=[].slice.call(document.querySelectorAll(a)).filter(function(e){return e===t})
null!==a&&0<l.length&&(this._selector=a,this._triggerArray.push(r))}this._parent=this._config.parent?this._getParent():null,this._config.parent||this._addAriaAndCollapsedClass(this._element,this._triggerArray),this._config.toggle&&this.toggle()}var n=t.prototype
return n.toggle=function(){e(this._element).hasClass(Q)?this.hide():this.show()},n.show=function(){var n,i,o=this
if(!(this._isTransitioning||e(this._element).hasClass(Q)||(this._parent&&0===(n=[].slice.call(this._parent.querySelectorAll(".show, .collapsing")).filter(function(t){return"string"==typeof o._config.parent?t.getAttribute("data-parent")===o._config.parent:t.classList.contains(V)})).length&&(n=null),n&&(i=e(n).not(this._selector).data(M))&&i._isTransitioning))){var r=e.Event(K.SHOW)
if(e(this._element).trigger(r),!r.isDefaultPrevented()){n&&(t._jQueryInterface.call(e(n).not(this._selector),"hide"),i||e(n).data(M,null))
var a=this._getDimension()
e(this._element).removeClass(V).addClass(Y),this._element.style[a]=0,this._triggerArray.length&&e(this._triggerArray).removeClass(z).attr("aria-expanded",!0),this.setTransitioning(!0)
var l="scroll"+(a[0].toUpperCase()+a.slice(1)),c=s.getTransitionDurationFromElement(this._element)
e(this._element).one(s.TRANSITION_END,function(){e(o._element).removeClass(Y).addClass(V).addClass(Q),o._element.style[a]="",o.setTransitioning(!1),e(o._element).trigger(K.SHOWN)}).emulateTransitionEnd(c),this._element.style[a]=this._element[l]+"px"}}},n.hide=function(){var t=this
if(!this._isTransitioning&&e(this._element).hasClass(Q)){var n=e.Event(K.HIDE)
if(e(this._element).trigger(n),!n.isDefaultPrevented()){var i=this._getDimension()
this._element.style[i]=this._element.getBoundingClientRect()[i]+"px",s.reflow(this._element),e(this._element).addClass(Y).removeClass(V).removeClass(Q)
var o=this._triggerArray.length
if(0<o)for(var r=0;r<o;r++){var a=this._triggerArray[r],l=s.getSelectorFromElement(a)
null!==l&&(e([].slice.call(document.querySelectorAll(l))).hasClass(Q)||e(a).addClass(z).attr("aria-expanded",!1))}this.setTransitioning(!0),this._element.style[i]=""
var c=s.getTransitionDurationFromElement(this._element)
e(this._element).one(s.TRANSITION_END,function(){t.setTransitioning(!1),e(t._element).removeClass(Y).addClass(V).trigger(K.HIDDEN)}).emulateTransitionEnd(c)}}},n.setTransitioning=function(t){this._isTransitioning=t},n.dispose=function(){e.removeData(this._element,M),this._config=null,this._parent=null,this._element=null,this._triggerArray=null,this._isTransitioning=null},n._getConfig=function(t){return(t=o({},B,t)).toggle=Boolean(t.toggle),s.typeCheckConfig(F,t,q),t},n._getDimension=function(){return e(this._element).hasClass("width")?"width":"height"},n._getParent=function(){var n,i=this
s.isElement(this._config.parent)?(n=this._config.parent,void 0!==this._config.parent.jquery&&(n=this._config.parent[0])):n=document.querySelector(this._config.parent)
var o='[data-toggle="collapse"][data-parent="'+this._config.parent+'"]',r=[].slice.call(n.querySelectorAll(o))
return e(r).each(function(e,n){i._addAriaAndCollapsedClass(t._getTargetFromElement(n),[n])}),n},n._addAriaAndCollapsedClass=function(t,n){var i=e(t).hasClass(Q)
n.length&&e(n).toggleClass(z,!i).attr("aria-expanded",i)},t._getTargetFromElement=function(t){var e=s.getSelectorFromElement(t)
return e?document.querySelector(e):null},t._jQueryInterface=function(n){return this.each(function(){var i=e(this),r=i.data(M),s=o({},B,i.data(),"object"==typeof n&&n?n:{})
if(!r&&s.toggle&&/show|hide/.test(n)&&(s.toggle=!1),r||(r=new t(this,s),i.data(M,r)),"string"==typeof n){if(void 0===r[n])throw new TypeError('No method named "'+n+'"')
r[n]()}})},i(t,null,[{key:"VERSION",get:function(){return"4.3.1"}},{key:"Default",get:function(){return B}}]),t}()
e(document).on(K.CLICK_DATA_API,X,function(t){"A"===t.currentTarget.tagName&&t.preventDefault()
var n=e(this),i=s.getSelectorFromElement(this),o=[].slice.call(document.querySelectorAll(i))
e(o).each(function(){var t=e(this),i=t.data(M)?"toggle":n.data()
G._jQueryInterface.call(t,i)})}),e.fn[F]=G._jQueryInterface,e.fn[F].Constructor=G,e.fn[F].noConflict=function(){return e.fn[F]=U,G._jQueryInterface}
for(var $="undefined"!=typeof window&&"undefined"!=typeof document,J=["Edge","Trident","Firefox"],Z=0,tt=0;tt<J.length;tt+=1)if($&&0<=navigator.userAgent.indexOf(J[tt])){Z=1
break}var et=$&&window.Promise?function(t){var e=!1
return function(){e||(e=!0,window.Promise.resolve().then(function(){e=!1,t()}))}}:function(t){var e=!1
return function(){e||(e=!0,setTimeout(function(){e=!1,t()},Z))}}
function nt(t){return t&&"[object Function]"==={}.toString.call(t)}function it(t,e){if(1!==t.nodeType)return[]
var n=t.ownerDocument.defaultView.getComputedStyle(t,null)
return e?n[e]:n}function ot(t){return"HTML"===t.nodeName?t:t.parentNode||t.host}function rt(t){if(!t)return document.body
switch(t.nodeName){case"HTML":case"BODY":return t.ownerDocument.body
case"#document":return t.body}var e=it(t),n=e.overflow,i=e.overflowX,o=e.overflowY
return/(auto|scroll|overlay)/.test(n+o+i)?t:rt(ot(t))}var st=$&&!(!window.MSInputMethodContext||!document.documentMode),at=$&&/MSIE 10/.test(navigator.userAgent)
function lt(t){return 11===t?st:10===t?at:st||at}function ct(t){if(!t)return document.documentElement
for(var e=lt(10)?document.body:null,n=t.offsetParent||null;n===e&&t.nextElementSibling;)n=(t=t.nextElementSibling).offsetParent
var i=n&&n.nodeName
return i&&"BODY"!==i&&"HTML"!==i?-1!==["TH","TD","TABLE"].indexOf(n.nodeName)&&"static"===it(n,"position")?ct(n):n:t?t.ownerDocument.documentElement:document.documentElement}function ht(t){return null!==t.parentNode?ht(t.parentNode):t}function ut(t,e){if(!(t&&t.nodeType&&e&&e.nodeType))return document.documentElement
var n=t.compareDocumentPosition(e)&Node.DOCUMENT_POSITION_FOLLOWING,i=n?t:e,o=n?e:t,r=document.createRange()
r.setStart(i,0),r.setEnd(o,0)
var s,a,l=r.commonAncestorContainer
if(t!==l&&e!==l||i.contains(o))return"BODY"===(a=(s=l).nodeName)||"HTML"!==a&&ct(s.firstElementChild)!==s?ct(l):l
var c=ht(t)
return c.host?ut(c.host,e):ut(t,ht(e).host)}function ft(t){var e="top"===(1<arguments.length&&void 0!==arguments[1]?arguments[1]:"top")?"scrollTop":"scrollLeft",n=t.nodeName
if("BODY"!==n&&"HTML"!==n)return t[e]
var i=t.ownerDocument.documentElement
return(t.ownerDocument.scrollingElement||i)[e]}function dt(t,e){var n="x"===e?"Left":"Top",i="Left"===n?"Right":"Bottom"
return parseFloat(t["border"+n+"Width"],10)+parseFloat(t["border"+i+"Width"],10)}function pt(t,e,n,i){return Math.max(e["offset"+t],e["scroll"+t],n["client"+t],n["offset"+t],n["scroll"+t],lt(10)?parseInt(n["offset"+t])+parseInt(i["margin"+("Height"===t?"Top":"Left")])+parseInt(i["margin"+("Height"===t?"Bottom":"Right")]):0)}function mt(t){var e=t.body,n=t.documentElement,i=lt(10)&&getComputedStyle(n)
return{height:pt("Height",e,n,i),width:pt("Width",e,n,i)}}var gt=function(){function t(t,e){for(var n=0;n<e.length;n++){var i=e[n]
i.enumerable=i.enumerable||!1,i.configurable=!0,"value"in i&&(i.writable=!0),Object.defineProperty(t,i.key,i)}}return function(e,n,i){return n&&t(e.prototype,n),i&&t(e,i),e}}(),_t=function(t,e,n){return e in t?Object.defineProperty(t,e,{value:n,enumerable:!0,configurable:!0,writable:!0}):t[e]=n,t},vt=Object.assign||function(t){for(var e=1;e<arguments.length;e++){var n=arguments[e]
for(var i in n)Object.prototype.hasOwnProperty.call(n,i)&&(t[i]=n[i])}return t}
function yt(t){return vt({},t,{right:t.left+t.width,bottom:t.top+t.height})}function Et(t){var e={}
try{if(lt(10)){e=t.getBoundingClientRect()
var n=ft(t,"top"),i=ft(t,"left")
e.top+=n,e.left+=i,e.bottom+=n,e.right+=i}else e=t.getBoundingClientRect()}catch(t){}var o={left:e.left,top:e.top,width:e.right-e.left,height:e.bottom-e.top},r="HTML"===t.nodeName?mt(t.ownerDocument):{},s=r.width||t.clientWidth||o.right-o.left,a=r.height||t.clientHeight||o.bottom-o.top,l=t.offsetWidth-s,c=t.offsetHeight-a
if(l||c){var h=it(t)
l-=dt(h,"x"),c-=dt(h,"y"),o.width-=l,o.height-=c}return yt(o)}function bt(t,e){var n=2<arguments.length&&void 0!==arguments[2]&&arguments[2],i=lt(10),o="HTML"===e.nodeName,r=Et(t),s=Et(e),a=rt(t),l=it(e),c=parseFloat(l.borderTopWidth,10),h=parseFloat(l.borderLeftWidth,10)
n&&o&&(s.top=Math.max(s.top,0),s.left=Math.max(s.left,0))
var u=yt({top:r.top-s.top-c,left:r.left-s.left-h,width:r.width,height:r.height})
if(u.marginTop=0,u.marginLeft=0,!i&&o){var f=parseFloat(l.marginTop,10),d=parseFloat(l.marginLeft,10)
u.top-=c-f,u.bottom-=c-f,u.left-=h-d,u.right-=h-d,u.marginTop=f,u.marginLeft=d}return(i&&!n?e.contains(a):e===a&&"BODY"!==a.nodeName)&&(u=function(t,e){var n=2<arguments.length&&void 0!==arguments[2]&&arguments[2],i=ft(e,"top"),o=ft(e,"left"),r=n?-1:1
return t.top+=i*r,t.bottom+=i*r,t.left+=o*r,t.right+=o*r,t}(u,e)),u}function wt(t){if(!t||!t.parentElement||lt())return document.documentElement
for(var e=t.parentElement;e&&"none"===it(e,"transform");)e=e.parentElement
return e||document.documentElement}function Ct(t,e,n,i){var o=4<arguments.length&&void 0!==arguments[4]&&arguments[4],r={top:0,left:0},s=o?wt(t):ut(t,e)
if("viewport"===i)r=function(t){var e=1<arguments.length&&void 0!==arguments[1]&&arguments[1],n=t.ownerDocument.documentElement,i=bt(t,n),o=Math.max(n.clientWidth,window.innerWidth||0),r=Math.max(n.clientHeight,window.innerHeight||0),s=e?0:ft(n),a=e?0:ft(n,"left")
return yt({top:s-i.top+i.marginTop,left:a-i.left+i.marginLeft,width:o,height:r})}(s,o)
else{var a=void 0
"scrollParent"===i?"BODY"===(a=rt(ot(e))).nodeName&&(a=t.ownerDocument.documentElement):a="window"===i?t.ownerDocument.documentElement:i
var l=bt(a,s,o)
if("HTML"!==a.nodeName||function t(e){var n=e.nodeName
if("BODY"===n||"HTML"===n)return!1
if("fixed"===it(e,"position"))return!0
var i=ot(e)
return!!i&&t(i)}(s))r=l
else{var c=mt(t.ownerDocument),h=c.height,u=c.width
r.top+=l.top-l.marginTop,r.bottom=h+l.top,r.left+=l.left-l.marginLeft,r.right=u+l.left}}var f="number"==typeof(n=n||0)
return r.left+=f?n:n.left||0,r.top+=f?n:n.top||0,r.right-=f?n:n.right||0,r.bottom-=f?n:n.bottom||0,r}function Tt(t,e,n,i,o){var r=5<arguments.length&&void 0!==arguments[5]?arguments[5]:0
if(-1===t.indexOf("auto"))return t
var s=Ct(n,i,r,o),a={top:{width:s.width,height:e.top-s.top},right:{width:s.right-e.right,height:s.height},bottom:{width:s.width,height:s.bottom-e.bottom},left:{width:e.left-s.left,height:s.height}},l=Object.keys(a).map(function(t){return vt({key:t},a[t],{area:(e=a[t],e.width*e.height)})
var e}).sort(function(t,e){return e.area-t.area}),c=l.filter(function(t){var e=t.width,i=t.height
return e>=n.clientWidth&&i>=n.clientHeight}),h=0<c.length?c[0].key:l[0].key,u=t.split("-")[1]
return h+(u?"-"+u:"")}function St(t,e,n){var i=3<arguments.length&&void 0!==arguments[3]?arguments[3]:null
return bt(n,i?wt(e):ut(e,n),i)}function Dt(t){var e=t.ownerDocument.defaultView.getComputedStyle(t),n=parseFloat(e.marginTop||0)+parseFloat(e.marginBottom||0),i=parseFloat(e.marginLeft||0)+parseFloat(e.marginRight||0)
return{width:t.offsetWidth+i,height:t.offsetHeight+n}}function It(t){var e={left:"right",right:"left",bottom:"top",top:"bottom"}
return t.replace(/left|right|bottom|top/g,function(t){return e[t]})}function At(t,e,n){n=n.split("-")[0]
var i=Dt(t),o={width:i.width,height:i.height},r=-1!==["right","left"].indexOf(n),s=r?"top":"left",a=r?"left":"top",l=r?"height":"width",c=r?"width":"height"
return o[s]=e[s]+e[l]/2-i[l]/2,o[a]=n===a?e[a]-i[c]:e[It(a)],o}function Ot(t,e){return Array.prototype.find?t.find(e):t.filter(e)[0]}function Nt(t,e,n){return(void 0===n?t:t.slice(0,function(t,e,n){if(Array.prototype.findIndex)return t.findIndex(function(t){return t[e]===n})
var i=Ot(t,function(t){return t[e]===n})
return t.indexOf(i)}(t,"name",n))).forEach(function(t){t.function&&console.warn("`modifier.function` is deprecated, use `modifier.fn`!")
var n=t.function||t.fn
t.enabled&&nt(n)&&(e.offsets.popper=yt(e.offsets.popper),e.offsets.reference=yt(e.offsets.reference),e=n(e,t))}),e}function kt(t,e){return t.some(function(t){var n=t.name
return t.enabled&&n===e})}function Lt(t){for(var e=[!1,"ms","Webkit","Moz","O"],n=t.charAt(0).toUpperCase()+t.slice(1),i=0;i<e.length;i++){var o=e[i],r=o?""+o+n:t
if(void 0!==document.body.style[r])return r}return null}function xt(t){var e=t.ownerDocument
return e?e.defaultView:window}function Pt(t){return""!==t&&!isNaN(parseFloat(t))&&isFinite(t)}function Ht(t,e){Object.keys(e).forEach(function(n){var i="";-1!==["width","height","top","right","bottom","left"].indexOf(n)&&Pt(e[n])&&(i="px"),t.style[n]=e[n]+i})}var jt=$&&/Firefox/i.test(navigator.userAgent)
function Rt(t,e,n){var i=Ot(t,function(t){return t.name===e}),o=!!i&&t.some(function(t){return t.name===n&&t.enabled&&t.order<i.order})
if(!o){var r="`"+e+"`",s="`"+n+"`"
console.warn(s+" modifier is required by "+r+" modifier in order to work, be sure to include it before "+r+"!")}return o}var Ft=["auto-start","auto","auto-end","top-start","top","top-end","right-start","right","right-end","bottom-end","bottom","bottom-start","left-end","left","left-start"],Mt=Ft.slice(3)
function Wt(t){var e=1<arguments.length&&void 0!==arguments[1]&&arguments[1],n=Mt.indexOf(t),i=Mt.slice(n+1).concat(Mt.slice(0,n))
return e?i.reverse():i}var Ut={placement:"bottom",positionFixed:!1,eventsEnabled:!0,removeOnDestroy:!1,onCreate:function(){},onUpdate:function(){},modifiers:{shift:{order:100,enabled:!0,fn:function(t){var e=t.placement,n=e.split("-")[0],i=e.split("-")[1]
if(i){var o=t.offsets,r=o.reference,s=o.popper,a=-1!==["bottom","top"].indexOf(n),l=a?"left":"top",c=a?"width":"height",h={start:_t({},l,r[l]),end:_t({},l,r[l]+r[c]-s[c])}
t.offsets.popper=vt({},s,h[i])}return t}},offset:{order:200,enabled:!0,fn:function(t,e){var n,i=e.offset,o=t.placement,r=t.offsets,s=r.popper,a=r.reference,l=o.split("-")[0]
return n=Pt(+i)?[+i,0]:function(t,e,n,i){var o=[0,0],r=-1!==["right","left"].indexOf(i),s=t.split(/(\+|\-)/).map(function(t){return t.trim()}),a=s.indexOf(Ot(s,function(t){return-1!==t.search(/,|\s/)}))
s[a]&&-1===s[a].indexOf(",")&&console.warn("Offsets separated by white space(s) are deprecated, use a comma (,) instead.")
var l=/\s*,\s*|\s+/,c=-1!==a?[s.slice(0,a).concat([s[a].split(l)[0]]),[s[a].split(l)[1]].concat(s.slice(a+1))]:[s]
return(c=c.map(function(t,i){var o=(1===i?!r:r)?"height":"width",s=!1
return t.reduce(function(t,e){return""===t[t.length-1]&&-1!==["+","-"].indexOf(e)?(t[t.length-1]=e,s=!0,t):s?(t[t.length-1]+=e,s=!1,t):t.concat(e)},[]).map(function(t){return function(t,e,n,i){var o=t.match(/((?:\-|\+)?\d*\.?\d*)(.*)/),r=+o[1],s=o[2]
if(!r)return t
if(0!==s.indexOf("%"))return"vh"!==s&&"vw"!==s?r:("vh"===s?Math.max(document.documentElement.clientHeight,window.innerHeight||0):Math.max(document.documentElement.clientWidth,window.innerWidth||0))/100*r
var a=void 0
switch(s){case"%p":a=n
break
case"%":case"%r":default:a=i}return yt(a)[e]/100*r}(t,o,e,n)})})).forEach(function(t,e){t.forEach(function(n,i){Pt(n)&&(o[e]+=n*("-"===t[i-1]?-1:1))})}),o}(i,s,a,l),"left"===l?(s.top+=n[0],s.left-=n[1]):"right"===l?(s.top+=n[0],s.left+=n[1]):"top"===l?(s.left+=n[0],s.top-=n[1]):"bottom"===l&&(s.left+=n[0],s.top+=n[1]),t.popper=s,t},offset:0},preventOverflow:{order:300,enabled:!0,fn:function(t,e){var n=e.boundariesElement||ct(t.instance.popper)
t.instance.reference===n&&(n=ct(n))
var i=Lt("transform"),o=t.instance.popper.style,r=o.top,s=o.left,a=o[i]
o.top="",o.left="",o[i]=""
var l=Ct(t.instance.popper,t.instance.reference,e.padding,n,t.positionFixed)
o.top=r,o.left=s,o[i]=a,e.boundaries=l
var c=e.priority,h=t.offsets.popper,u={primary:function(t){var n=h[t]
return h[t]<l[t]&&!e.escapeWithReference&&(n=Math.max(h[t],l[t])),_t({},t,n)},secondary:function(t){var n="right"===t?"left":"top",i=h[n]
return h[t]>l[t]&&!e.escapeWithReference&&(i=Math.min(h[n],l[t]-("right"===t?h.width:h.height))),_t({},n,i)}}
return c.forEach(function(t){var e=-1!==["left","top"].indexOf(t)?"primary":"secondary"
h=vt({},h,u[e](t))}),t.offsets.popper=h,t},priority:["left","right","top","bottom"],padding:5,boundariesElement:"scrollParent"},keepTogether:{order:400,enabled:!0,fn:function(t){var e=t.offsets,n=e.popper,i=e.reference,o=t.placement.split("-")[0],r=Math.floor,s=-1!==["top","bottom"].indexOf(o),a=s?"right":"bottom",l=s?"left":"top",c=s?"width":"height"
return n[a]<r(i[l])&&(t.offsets.popper[l]=r(i[l])-n[c]),n[l]>r(i[a])&&(t.offsets.popper[l]=r(i[a])),t}},arrow:{order:500,enabled:!0,fn:function(t,e){var n
if(!Rt(t.instance.modifiers,"arrow","keepTogether"))return t
var i=e.element
if("string"==typeof i){if(!(i=t.instance.popper.querySelector(i)))return t}else if(!t.instance.popper.contains(i))return console.warn("WARNING: `arrow.element` must be child of its popper element!"),t
var o=t.placement.split("-")[0],r=t.offsets,s=r.popper,a=r.reference,l=-1!==["left","right"].indexOf(o),c=l?"height":"width",h=l?"Top":"Left",u=h.toLowerCase(),f=l?"left":"top",d=l?"bottom":"right",p=Dt(i)[c]
a[d]-p<s[u]&&(t.offsets.popper[u]-=s[u]-(a[d]-p)),a[u]+p>s[d]&&(t.offsets.popper[u]+=a[u]+p-s[d]),t.offsets.popper=yt(t.offsets.popper)
var m=a[u]+a[c]/2-p/2,g=it(t.instance.popper),_=parseFloat(g["margin"+h],10),v=parseFloat(g["border"+h+"Width"],10),y=m-t.offsets.popper[u]-_-v
return y=Math.max(Math.min(s[c]-p,y),0),t.arrowElement=i,t.offsets.arrow=(_t(n={},u,Math.round(y)),_t(n,f,""),n),t},element:"[x-arrow]"},flip:{order:600,enabled:!0,fn:function(t,e){if(kt(t.instance.modifiers,"inner"))return t
if(t.flipped&&t.placement===t.originalPlacement)return t
var n=Ct(t.instance.popper,t.instance.reference,e.padding,e.boundariesElement,t.positionFixed),i=t.placement.split("-")[0],o=It(i),r=t.placement.split("-")[1]||"",s=[]
switch(e.behavior){case"flip":s=[i,o]
break
case"clockwise":s=Wt(i)
break
case"counterclockwise":s=Wt(i,!0)
break
default:s=e.behavior}return s.forEach(function(a,l){if(i!==a||s.length===l+1)return t
i=t.placement.split("-")[0],o=It(i)
var c,h=t.offsets.popper,u=t.offsets.reference,f=Math.floor,d="left"===i&&f(h.right)>f(u.left)||"right"===i&&f(h.left)<f(u.right)||"top"===i&&f(h.bottom)>f(u.top)||"bottom"===i&&f(h.top)<f(u.bottom),p=f(h.left)<f(n.left),m=f(h.right)>f(n.right),g=f(h.top)<f(n.top),_=f(h.bottom)>f(n.bottom),v="left"===i&&p||"right"===i&&m||"top"===i&&g||"bottom"===i&&_,y=-1!==["top","bottom"].indexOf(i),E=!!e.flipVariations&&(y&&"start"===r&&p||y&&"end"===r&&m||!y&&"start"===r&&g||!y&&"end"===r&&_);(d||v||E)&&(t.flipped=!0,(d||v)&&(i=s[l+1]),E&&(r="end"===(c=r)?"start":"start"===c?"end":c),t.placement=i+(r?"-"+r:""),t.offsets.popper=vt({},t.offsets.popper,At(t.instance.popper,t.offsets.reference,t.placement)),t=Nt(t.instance.modifiers,t,"flip"))}),t},behavior:"flip",padding:5,boundariesElement:"viewport"},inner:{order:700,enabled:!1,fn:function(t){var e=t.placement,n=e.split("-")[0],i=t.offsets,o=i.popper,r=i.reference,s=-1!==["left","right"].indexOf(n),a=-1===["top","left"].indexOf(n)
return o[s?"left":"top"]=r[n]-(a?o[s?"width":"height"]:0),t.placement=It(e),t.offsets.popper=yt(o),t}},hide:{order:800,enabled:!0,fn:function(t){if(!Rt(t.instance.modifiers,"hide","preventOverflow"))return t
var e=t.offsets.reference,n=Ot(t.instance.modifiers,function(t){return"preventOverflow"===t.name}).boundaries
if(e.bottom<n.top||e.left>n.right||e.top>n.bottom||e.right<n.left){if(!0===t.hide)return t
t.hide=!0,t.attributes["x-out-of-boundaries"]=""}else{if(!1===t.hide)return t
t.hide=!1,t.attributes["x-out-of-boundaries"]=!1}return t}},computeStyle:{order:850,enabled:!0,fn:function(t,e){var n=e.x,i=e.y,o=t.offsets.popper,r=Ot(t.instance.modifiers,function(t){return"applyStyle"===t.name}).gpuAcceleration
void 0!==r&&console.warn("WARNING: `gpuAcceleration` option moved to `computeStyle` modifier and will not be supported in future versions of Popper.js!")
var s,a,l,c,h,u,f,d,p,m,g,_,v,y,E,b,w=void 0!==r?r:e.gpuAcceleration,C=ct(t.instance.popper),T=Et(C),S={position:o.position},D=(s=t,a=window.devicePixelRatio<2||!jt,c=(l=s.offsets).popper,h=l.reference,u=Math.round,f=Math.floor,d=function(t){return t},p=u(h.width),m=u(c.width),g=-1!==["left","right"].indexOf(s.placement),_=-1!==s.placement.indexOf("-"),y=a?u:d,{left:(v=a?g||_||p%2==m%2?u:f:d)(p%2==1&&m%2==1&&!_&&a?c.left-1:c.left),top:y(c.top),bottom:y(c.bottom),right:v(c.right)}),I="bottom"===n?"top":"bottom",A="right"===i?"left":"right",O=Lt("transform")
if(b="bottom"===I?"HTML"===C.nodeName?-C.clientHeight+D.bottom:-T.height+D.bottom:D.top,E="right"===A?"HTML"===C.nodeName?-C.clientWidth+D.right:-T.width+D.right:D.left,w&&O)S[O]="translate3d("+E+"px, "+b+"px, 0)",S[I]=0,S[A]=0,S.willChange="transform"
else{var N="bottom"===I?-1:1,k="right"===A?-1:1
S[I]=b*N,S[A]=E*k,S.willChange=I+", "+A}var L={"x-placement":t.placement}
return t.attributes=vt({},L,t.attributes),t.styles=vt({},S,t.styles),t.arrowStyles=vt({},t.offsets.arrow,t.arrowStyles),t},gpuAcceleration:!0,x:"bottom",y:"right"},applyStyle:{order:900,enabled:!0,fn:function(t){var e,n
return Ht(t.instance.popper,t.styles),e=t.instance.popper,n=t.attributes,Object.keys(n).forEach(function(t){!1!==n[t]?e.setAttribute(t,n[t]):e.removeAttribute(t)}),t.arrowElement&&Object.keys(t.arrowStyles).length&&Ht(t.arrowElement,t.arrowStyles),t},onLoad:function(t,e,n,i,o){var r=St(o,e,t,n.positionFixed),s=Tt(n.placement,r,e,t,n.modifiers.flip.boundariesElement,n.modifiers.flip.padding)
return e.setAttribute("x-placement",s),Ht(e,{position:n.positionFixed?"fixed":"absolute"}),n},gpuAcceleration:void 0}}},Bt=function(){function t(e,n){var i=this,o=2<arguments.length&&void 0!==arguments[2]?arguments[2]:{}
!function(e,n){if(!(e instanceof t))throw new TypeError("Cannot call a class as a function")}(this),this.scheduleUpdate=function(){return requestAnimationFrame(i.update)},this.update=et(this.update.bind(this)),this.options=vt({},t.Defaults,o),this.state={isDestroyed:!1,isCreated:!1,scrollParents:[]},this.reference=e&&e.jquery?e[0]:e,this.popper=n&&n.jquery?n[0]:n,this.options.modifiers={},Object.keys(vt({},t.Defaults.modifiers,o.modifiers)).forEach(function(e){i.options.modifiers[e]=vt({},t.Defaults.modifiers[e]||{},o.modifiers?o.modifiers[e]:{})}),this.modifiers=Object.keys(this.options.modifiers).map(function(t){return vt({name:t},i.options.modifiers[t])}).sort(function(t,e){return t.order-e.order}),this.modifiers.forEach(function(t){t.enabled&&nt(t.onLoad)&&t.onLoad(i.reference,i.popper,i.options,t,i.state)}),this.update()
var r=this.options.eventsEnabled
r&&this.enableEventListeners(),this.state.eventsEnabled=r}return gt(t,[{key:"update",value:function(){return function(){if(!this.state.isDestroyed){var t={instance:this,styles:{},arrowStyles:{},attributes:{},flipped:!1,offsets:{}}
t.offsets.reference=St(this.state,this.popper,this.reference,this.options.positionFixed),t.placement=Tt(this.options.placement,t.offsets.reference,this.popper,this.reference,this.options.modifiers.flip.boundariesElement,this.options.modifiers.flip.padding),t.originalPlacement=t.placement,t.positionFixed=this.options.positionFixed,t.offsets.popper=At(this.popper,t.offsets.reference,t.placement),t.offsets.popper.position=this.options.positionFixed?"fixed":"absolute",t=Nt(this.modifiers,t),this.state.isCreated?this.options.onUpdate(t):(this.state.isCreated=!0,this.options.onCreate(t))}}.call(this)}},{key:"destroy",value:function(){return function(){return this.state.isDestroyed=!0,kt(this.modifiers,"applyStyle")&&(this.popper.removeAttribute("x-placement"),this.popper.style.position="",this.popper.style.top="",this.popper.style.left="",this.popper.style.right="",this.popper.style.bottom="",this.popper.style.willChange="",this.popper.style[Lt("transform")]=""),this.disableEventListeners(),this.options.removeOnDestroy&&this.popper.parentNode.removeChild(this.popper),this}.call(this)}},{key:"enableEventListeners",value:function(){return function(){this.state.eventsEnabled||(this.state=function(t,e,n,i){n.updateBound=i,xt(t).addEventListener("resize",n.updateBound,{passive:!0})
var o=rt(t)
return function t(e,n,i,o){var r="BODY"===e.nodeName,s=r?e.ownerDocument.defaultView:e
s.addEventListener(n,i,{passive:!0}),r||t(rt(s.parentNode),n,i,o),o.push(s)}(o,"scroll",n.updateBound,n.scrollParents),n.scrollElement=o,n.eventsEnabled=!0,n}(this.reference,this.options,this.state,this.scheduleUpdate))}.call(this)}},{key:"disableEventListeners",value:function(){return function(){var t,e
this.state.eventsEnabled&&(cancelAnimationFrame(this.scheduleUpdate),this.state=(t=this.reference,e=this.state,xt(t).removeEventListener("resize",e.updateBound),e.scrollParents.forEach(function(t){t.removeEventListener("scroll",e.updateBound)}),e.updateBound=null,e.scrollParents=[],e.scrollElement=null,e.eventsEnabled=!1,e))}.call(this)}}]),t}()
Bt.Utils=("undefined"!=typeof window?window:global).PopperUtils,Bt.placements=Ft,Bt.Defaults=Ut
var qt="dropdown",Kt="bs.dropdown",Qt="."+Kt,Vt=".data-api",Yt=e.fn[qt],zt=new RegExp("38|40|27"),Xt={HIDE:"hide"+Qt,HIDDEN:"hidden"+Qt,SHOW:"show"+Qt,SHOWN:"shown"+Qt,CLICK:"click"+Qt,CLICK_DATA_API:"click"+Qt+Vt,KEYDOWN_DATA_API:"keydown"+Qt+Vt,KEYUP_DATA_API:"keyup"+Qt+Vt},Gt="disabled",$t="show",Jt="dropdown-menu-right",Zt='[data-toggle="dropdown"]',te=".dropdown-menu",ee={offset:0,flip:!0,boundary:"scrollParent",reference:"toggle",display:"dynamic"},ne={offset:"(number|string|function)",flip:"boolean",boundary:"(string|element)",reference:"(string|element)",display:"string"},ie=function(){function t(t,e){this._element=t,this._popper=null,this._config=this._getConfig(e),this._menu=this._getMenuElement(),this._inNavbar=this._detectNavbar(),this._addEventListeners()}var n=t.prototype
return n.toggle=function(){if(!this._element.disabled&&!e(this._element).hasClass(Gt)){var n=t._getParentFromElement(this._element),i=e(this._menu).hasClass($t)
if(t._clearMenus(),!i){var o={relatedTarget:this._element},r=e.Event(Xt.SHOW,o)
if(e(n).trigger(r),!r.isDefaultPrevented()){if(!this._inNavbar){if(void 0===Bt)throw new TypeError("Bootstrap's dropdowns require Popper.js (https://popper.js.org/)")
var a=this._element
"parent"===this._config.reference?a=n:s.isElement(this._config.reference)&&(a=this._config.reference,void 0!==this._config.reference.jquery&&(a=this._config.reference[0])),"scrollParent"!==this._config.boundary&&e(n).addClass("position-static"),this._popper=new Bt(a,this._menu,this._getPopperConfig())}"ontouchstart"in document.documentElement&&0===e(n).closest(".navbar-nav").length&&e(document.body).children().on("mouseover",null,e.noop),this._element.focus(),this._element.setAttribute("aria-expanded",!0),e(this._menu).toggleClass($t),e(n).toggleClass($t).trigger(e.Event(Xt.SHOWN,o))}}}},n.show=function(){if(!(this._element.disabled||e(this._element).hasClass(Gt)||e(this._menu).hasClass($t))){var n={relatedTarget:this._element},i=e.Event(Xt.SHOW,n),o=t._getParentFromElement(this._element)
e(o).trigger(i),i.isDefaultPrevented()||(e(this._menu).toggleClass($t),e(o).toggleClass($t).trigger(e.Event(Xt.SHOWN,n)))}},n.hide=function(){if(!this._element.disabled&&!e(this._element).hasClass(Gt)&&e(this._menu).hasClass($t)){var n={relatedTarget:this._element},i=e.Event(Xt.HIDE,n),o=t._getParentFromElement(this._element)
e(o).trigger(i),i.isDefaultPrevented()||(e(this._menu).toggleClass($t),e(o).toggleClass($t).trigger(e.Event(Xt.HIDDEN,n)))}},n.dispose=function(){e.removeData(this._element,Kt),e(this._element).off(Qt),this._element=null,(this._menu=null)!==this._popper&&(this._popper.destroy(),this._popper=null)},n.update=function(){this._inNavbar=this._detectNavbar(),null!==this._popper&&this._popper.scheduleUpdate()},n._addEventListeners=function(){var t=this
e(this._element).on(Xt.CLICK,function(e){e.preventDefault(),e.stopPropagation(),t.toggle()})},n._getConfig=function(t){return t=o({},this.constructor.Default,e(this._element).data(),t),s.typeCheckConfig(qt,t,this.constructor.DefaultType),t},n._getMenuElement=function(){if(!this._menu){var e=t._getParentFromElement(this._element)
e&&(this._menu=e.querySelector(te))}return this._menu},n._getPlacement=function(){var t=e(this._element.parentNode),n="bottom-start"
return t.hasClass("dropup")?(n="top-start",e(this._menu).hasClass(Jt)&&(n="top-end")):t.hasClass("dropright")?n="right-start":t.hasClass("dropleft")?n="left-start":e(this._menu).hasClass(Jt)&&(n="bottom-end"),n},n._detectNavbar=function(){return 0<e(this._element).closest(".navbar").length},n._getOffset=function(){var t=this,e={}
return"function"==typeof this._config.offset?e.fn=function(e){return e.offsets=o({},e.offsets,t._config.offset(e.offsets,t._element)||{}),e}:e.offset=this._config.offset,e},n._getPopperConfig=function(){var t={placement:this._getPlacement(),modifiers:{offset:this._getOffset(),flip:{enabled:this._config.flip},preventOverflow:{boundariesElement:this._config.boundary}}}
return"static"===this._config.display&&(t.modifiers.applyStyle={enabled:!1}),t},t._jQueryInterface=function(n){return this.each(function(){var i=e(this).data(Kt)
if(i||(i=new t(this,"object"==typeof n?n:null),e(this).data(Kt,i)),"string"==typeof n){if(void 0===i[n])throw new TypeError('No method named "'+n+'"')
i[n]()}})},t._clearMenus=function(n){if(!n||3!==n.which&&("keyup"!==n.type||9===n.which))for(var i=[].slice.call(document.querySelectorAll(Zt)),o=0,r=i.length;o<r;o++){var s=t._getParentFromElement(i[o]),a=e(i[o]).data(Kt),l={relatedTarget:i[o]}
if(n&&"click"===n.type&&(l.clickEvent=n),a){var c=a._menu
if(e(s).hasClass($t)&&!(n&&("click"===n.type&&/input|textarea/i.test(n.target.tagName)||"keyup"===n.type&&9===n.which)&&e.contains(s,n.target))){var h=e.Event(Xt.HIDE,l)
e(s).trigger(h),h.isDefaultPrevented()||("ontouchstart"in document.documentElement&&e(document.body).children().off("mouseover",null,e.noop),i[o].setAttribute("aria-expanded","false"),e(c).removeClass($t),e(s).removeClass($t).trigger(e.Event(Xt.HIDDEN,l)))}}}},t._getParentFromElement=function(t){var e,n=s.getSelectorFromElement(t)
return n&&(e=document.querySelector(n)),e||t.parentNode},t._dataApiKeydownHandler=function(n){if((/input|textarea/i.test(n.target.tagName)?!(32===n.which||27!==n.which&&(40!==n.which&&38!==n.which||e(n.target).closest(te).length)):zt.test(n.which))&&(n.preventDefault(),n.stopPropagation(),!this.disabled&&!e(this).hasClass(Gt))){var i=t._getParentFromElement(this),o=e(i).hasClass($t)
if(o&&(!o||27!==n.which&&32!==n.which)){var r=[].slice.call(i.querySelectorAll(".dropdown-menu .dropdown-item:not(.disabled):not(:disabled)"))
if(0!==r.length){var s=r.indexOf(n.target)
38===n.which&&0<s&&s--,40===n.which&&s<r.length-1&&s++,s<0&&(s=0),r[s].focus()}}else{if(27===n.which){var a=i.querySelector(Zt)
e(a).trigger("focus")}e(this).trigger("click")}}},i(t,null,[{key:"VERSION",get:function(){return"4.3.1"}},{key:"Default",get:function(){return ee}},{key:"DefaultType",get:function(){return ne}}]),t}()
e(document).on(Xt.KEYDOWN_DATA_API,Zt,ie._dataApiKeydownHandler).on(Xt.KEYDOWN_DATA_API,te,ie._dataApiKeydownHandler).on(Xt.CLICK_DATA_API+" "+Xt.KEYUP_DATA_API,ie._clearMenus).on(Xt.CLICK_DATA_API,Zt,function(t){t.preventDefault(),t.stopPropagation(),ie._jQueryInterface.call(e(this),"toggle")}).on(Xt.CLICK_DATA_API,".dropdown form",function(t){t.stopPropagation()}),e.fn[qt]=ie._jQueryInterface,e.fn[qt].Constructor=ie,e.fn[qt].noConflict=function(){return e.fn[qt]=Yt,ie._jQueryInterface}
var oe="modal",re="bs.modal",se="."+re,ae=e.fn[oe],le={backdrop:!0,keyboard:!0,focus:!0,show:!0},ce={backdrop:"(boolean|string)",keyboard:"boolean",focus:"boolean",show:"boolean"},he={HIDE:"hide"+se,HIDDEN:"hidden"+se,SHOW:"show"+se,SHOWN:"shown"+se,FOCUSIN:"focusin"+se,RESIZE:"resize"+se,CLICK_DISMISS:"click.dismiss"+se,KEYDOWN_DISMISS:"keydown.dismiss"+se,MOUSEUP_DISMISS:"mouseup.dismiss"+se,MOUSEDOWN_DISMISS:"mousedown.dismiss"+se,CLICK_DATA_API:"click"+se+".data-api"},ue="modal-open",fe="fade",de="show",pe=".modal-dialog",me=".fixed-top, .fixed-bottom, .is-fixed, .sticky-top",ge=".sticky-top",_e=function(){function t(t,e){this._config=this._getConfig(e),this._element=t,this._dialog=t.querySelector(pe),this._backdrop=null,this._isShown=!1,this._isBodyOverflowing=!1,this._ignoreBackdropClick=!1,this._isTransitioning=!1,this._scrollbarWidth=0}var n=t.prototype
return n.toggle=function(t){return this._isShown?this.hide():this.show(t)},n.show=function(t){var n=this
if(!this._isShown&&!this._isTransitioning){e(this._element).hasClass(fe)&&(this._isTransitioning=!0)
var i=e.Event(he.SHOW,{relatedTarget:t})
e(this._element).trigger(i),this._isShown||i.isDefaultPrevented()||(this._isShown=!0,this._checkScrollbar(),this._setScrollbar(),this._adjustDialog(),this._setEscapeEvent(),this._setResizeEvent(),e(this._element).on(he.CLICK_DISMISS,'[data-dismiss="modal"]',function(t){return n.hide(t)}),e(this._dialog).on(he.MOUSEDOWN_DISMISS,function(){e(n._element).one(he.MOUSEUP_DISMISS,function(t){e(t.target).is(n._element)&&(n._ignoreBackdropClick=!0)})}),this._showBackdrop(function(){return n._showElement(t)}))}},n.hide=function(t){var n=this
if(t&&t.preventDefault(),this._isShown&&!this._isTransitioning){var i=e.Event(he.HIDE)
if(e(this._element).trigger(i),this._isShown&&!i.isDefaultPrevented()){this._isShown=!1
var o=e(this._element).hasClass(fe)
if(o&&(this._isTransitioning=!0),this._setEscapeEvent(),this._setResizeEvent(),e(document).off(he.FOCUSIN),e(this._element).removeClass(de),e(this._element).off(he.CLICK_DISMISS),e(this._dialog).off(he.MOUSEDOWN_DISMISS),o){var r=s.getTransitionDurationFromElement(this._element)
e(this._element).one(s.TRANSITION_END,function(t){return n._hideModal(t)}).emulateTransitionEnd(r)}else this._hideModal()}}},n.dispose=function(){[window,this._element,this._dialog].forEach(function(t){return e(t).off(se)}),e(document).off(he.FOCUSIN),e.removeData(this._element,re),this._config=null,this._element=null,this._dialog=null,this._backdrop=null,this._isShown=null,this._isBodyOverflowing=null,this._ignoreBackdropClick=null,this._isTransitioning=null,this._scrollbarWidth=null},n.handleUpdate=function(){this._adjustDialog()},n._getConfig=function(t){return t=o({},le,t),s.typeCheckConfig(oe,t,ce),t},n._showElement=function(t){var n=this,i=e(this._element).hasClass(fe)
this._element.parentNode&&this._element.parentNode.nodeType===Node.ELEMENT_NODE||document.body.appendChild(this._element),this._element.style.display="block",this._element.removeAttribute("aria-hidden"),this._element.setAttribute("aria-modal",!0),e(this._dialog).hasClass("modal-dialog-scrollable")?this._dialog.querySelector(".modal-body").scrollTop=0:this._element.scrollTop=0,i&&s.reflow(this._element),e(this._element).addClass(de),this._config.focus&&this._enforceFocus()
var o=e.Event(he.SHOWN,{relatedTarget:t}),r=function(){n._config.focus&&n._element.focus(),n._isTransitioning=!1,e(n._element).trigger(o)}
if(i){var a=s.getTransitionDurationFromElement(this._dialog)
e(this._dialog).one(s.TRANSITION_END,r).emulateTransitionEnd(a)}else r()},n._enforceFocus=function(){var t=this
e(document).off(he.FOCUSIN).on(he.FOCUSIN,function(n){document!==n.target&&t._element!==n.target&&0===e(t._element).has(n.target).length&&t._element.focus()})},n._setEscapeEvent=function(){var t=this
this._isShown&&this._config.keyboard?e(this._element).on(he.KEYDOWN_DISMISS,function(e){27===e.which&&(e.preventDefault(),t.hide())}):this._isShown||e(this._element).off(he.KEYDOWN_DISMISS)},n._setResizeEvent=function(){var t=this
this._isShown?e(window).on(he.RESIZE,function(e){return t.handleUpdate(e)}):e(window).off(he.RESIZE)},n._hideModal=function(){var t=this
this._element.style.display="none",this._element.setAttribute("aria-hidden",!0),this._element.removeAttribute("aria-modal"),this._isTransitioning=!1,this._showBackdrop(function(){e(document.body).removeClass(ue),t._resetAdjustments(),t._resetScrollbar(),e(t._element).trigger(he.HIDDEN)})},n._removeBackdrop=function(){this._backdrop&&(e(this._backdrop).remove(),this._backdrop=null)},n._showBackdrop=function(t){var n=this,i=e(this._element).hasClass(fe)?fe:""
if(this._isShown&&this._config.backdrop){if(this._backdrop=document.createElement("div"),this._backdrop.className="modal-backdrop",i&&this._backdrop.classList.add(i),e(this._backdrop).appendTo(document.body),e(this._element).on(he.CLICK_DISMISS,function(t){n._ignoreBackdropClick?n._ignoreBackdropClick=!1:t.target===t.currentTarget&&("static"===n._config.backdrop?n._element.focus():n.hide())}),i&&s.reflow(this._backdrop),e(this._backdrop).addClass(de),!t)return
if(!i)return void t()
var o=s.getTransitionDurationFromElement(this._backdrop)
e(this._backdrop).one(s.TRANSITION_END,t).emulateTransitionEnd(o)}else if(!this._isShown&&this._backdrop){e(this._backdrop).removeClass(de)
var r=function(){n._removeBackdrop(),t&&t()}
if(e(this._element).hasClass(fe)){var a=s.getTransitionDurationFromElement(this._backdrop)
e(this._backdrop).one(s.TRANSITION_END,r).emulateTransitionEnd(a)}else r()}else t&&t()},n._adjustDialog=function(){var t=this._element.scrollHeight>document.documentElement.clientHeight
!this._isBodyOverflowing&&t&&(this._element.style.paddingLeft=this._scrollbarWidth+"px"),this._isBodyOverflowing&&!t&&(this._element.style.paddingRight=this._scrollbarWidth+"px")},n._resetAdjustments=function(){this._element.style.paddingLeft="",this._element.style.paddingRight=""},n._checkScrollbar=function(){var t=document.body.getBoundingClientRect()
this._isBodyOverflowing=t.left+t.right<window.innerWidth,this._scrollbarWidth=this._getScrollbarWidth()},n._setScrollbar=function(){var t=this
if(this._isBodyOverflowing){var n=[].slice.call(document.querySelectorAll(me)),i=[].slice.call(document.querySelectorAll(ge))
e(n).each(function(n,i){var o=i.style.paddingRight,r=e(i).css("padding-right")
e(i).data("padding-right",o).css("padding-right",parseFloat(r)+t._scrollbarWidth+"px")}),e(i).each(function(n,i){var o=i.style.marginRight,r=e(i).css("margin-right")
e(i).data("margin-right",o).css("margin-right",parseFloat(r)-t._scrollbarWidth+"px")})
var o=document.body.style.paddingRight,r=e(document.body).css("padding-right")
e(document.body).data("padding-right",o).css("padding-right",parseFloat(r)+this._scrollbarWidth+"px")}e(document.body).addClass(ue)},n._resetScrollbar=function(){var t=[].slice.call(document.querySelectorAll(me))
e(t).each(function(t,n){var i=e(n).data("padding-right")
e(n).removeData("padding-right"),n.style.paddingRight=i||""})
var n=[].slice.call(document.querySelectorAll(""+ge))
e(n).each(function(t,n){var i=e(n).data("margin-right")
void 0!==i&&e(n).css("margin-right",i).removeData("margin-right")})
var i=e(document.body).data("padding-right")
e(document.body).removeData("padding-right"),document.body.style.paddingRight=i||""},n._getScrollbarWidth=function(){var t=document.createElement("div")
t.className="modal-scrollbar-measure",document.body.appendChild(t)
var e=t.getBoundingClientRect().width-t.clientWidth
return document.body.removeChild(t),e},t._jQueryInterface=function(n,i){return this.each(function(){var r=e(this).data(re),s=o({},le,e(this).data(),"object"==typeof n&&n?n:{})
if(r||(r=new t(this,s),e(this).data(re,r)),"string"==typeof n){if(void 0===r[n])throw new TypeError('No method named "'+n+'"')
r[n](i)}else s.show&&r.show(i)})},i(t,null,[{key:"VERSION",get:function(){return"4.3.1"}},{key:"Default",get:function(){return le}}]),t}()
e(document).on(he.CLICK_DATA_API,'[data-toggle="modal"]',function(t){var n,i=this,r=s.getSelectorFromElement(this)
r&&(n=document.querySelector(r))
var a=e(n).data(re)?"toggle":o({},e(n).data(),e(this).data())
"A"!==this.tagName&&"AREA"!==this.tagName||t.preventDefault()
var l=e(n).one(he.SHOW,function(t){t.isDefaultPrevented()||l.one(he.HIDDEN,function(){e(i).is(":visible")&&i.focus()})})
_e._jQueryInterface.call(e(n),a,this)}),e.fn[oe]=_e._jQueryInterface,e.fn[oe].Constructor=_e,e.fn[oe].noConflict=function(){return e.fn[oe]=ae,_e._jQueryInterface}
var ve=["background","cite","href","itemtype","longdesc","poster","src","xlink:href"],ye=/^(?:(?:https?|mailto|ftp|tel|file):|[^&:\/?#]*(?:[\/?#]|$))/gi,Ee=/^data:(?:image\/(?:bmp|gif|jpeg|jpg|png|tiff|webp)|video\/(?:mpeg|mp4|ogg|webm)|audio\/(?:mp3|oga|ogg|opus));base64,[a-z0-9+\/]+=*$/i
function be(t,e,n){if(0===t.length)return t
if(n&&"function"==typeof n)return n(t)
for(var i=(new window.DOMParser).parseFromString(t,"text/html"),o=Object.keys(e),r=[].slice.call(i.body.querySelectorAll("*")),s=function(t,n){var i=r[t],s=i.nodeName.toLowerCase()
if(-1===o.indexOf(i.nodeName.toLowerCase()))return i.parentNode.removeChild(i),"continue"
var a=[].slice.call(i.attributes),l=[].concat(e["*"]||[],e[s]||[])
a.forEach(function(t){(function(t,e){var n=t.nodeName.toLowerCase()
if(-1!==e.indexOf(n))return-1===ve.indexOf(n)||Boolean(t.nodeValue.match(ye)||t.nodeValue.match(Ee))
for(var i=e.filter(function(t){return t instanceof RegExp}),o=0,r=i.length;o<r;o++)if(n.match(i[o]))return!0
return!1})(t,l)||i.removeAttribute(t.nodeName)})},a=0,l=r.length;a<l;a++)s(a)
return i.body.innerHTML}var we="tooltip",Ce="bs.tooltip",Te="."+Ce,Se=e.fn[we],De="bs-tooltip",Ie=new RegExp("(^|\\s)"+De+"\\S+","g"),Ae=["sanitize","whiteList","sanitizeFn"],Oe={animation:"boolean",template:"string",title:"(string|element|function)",trigger:"string",delay:"(number|object)",html:"boolean",selector:"(string|boolean)",placement:"(string|function)",offset:"(number|string|function)",container:"(string|element|boolean)",fallbackPlacement:"(string|array)",boundary:"(string|element)",sanitize:"boolean",sanitizeFn:"(null|function)",whiteList:"object"},Ne={AUTO:"auto",TOP:"top",RIGHT:"right",BOTTOM:"bottom",LEFT:"left"},ke={animation:!0,template:'<div class="tooltip" role="tooltip"><div class="arrow"></div><div class="tooltip-inner"></div></div>',trigger:"hover focus",title:"",delay:0,html:!1,selector:!1,placement:"top",offset:0,container:!1,fallbackPlacement:"flip",boundary:"scrollParent",sanitize:!0,sanitizeFn:null,whiteList:{"*":["class","dir","id","lang","role",/^aria-[\w-]*$/i],a:["target","href","title","rel"],area:[],b:[],br:[],col:[],code:[],div:[],em:[],hr:[],h1:[],h2:[],h3:[],h4:[],h5:[],h6:[],i:[],img:["src","alt","title","width","height"],li:[],ol:[],p:[],pre:[],s:[],small:[],span:[],sub:[],sup:[],strong:[],u:[],ul:[]}},Le="show",xe={HIDE:"hide"+Te,HIDDEN:"hidden"+Te,SHOW:"show"+Te,SHOWN:"shown"+Te,INSERTED:"inserted"+Te,CLICK:"click"+Te,FOCUSIN:"focusin"+Te,FOCUSOUT:"focusout"+Te,MOUSEENTER:"mouseenter"+Te,MOUSELEAVE:"mouseleave"+Te},Pe="fade",He="show",je="hover",Re="focus",Fe=function(){function t(t,e){if(void 0===Bt)throw new TypeError("Bootstrap's tooltips require Popper.js (https://popper.js.org/)")
this._isEnabled=!0,this._timeout=0,this._hoverState="",this._activeTrigger={},this._popper=null,this.element=t,this.config=this._getConfig(e),this.tip=null,this._setListeners()}var n=t.prototype
return n.enable=function(){this._isEnabled=!0},n.disable=function(){this._isEnabled=!1},n.toggleEnabled=function(){this._isEnabled=!this._isEnabled},n.toggle=function(t){if(this._isEnabled)if(t){var n=this.constructor.DATA_KEY,i=e(t.currentTarget).data(n)
i||(i=new this.constructor(t.currentTarget,this._getDelegateConfig()),e(t.currentTarget).data(n,i)),i._activeTrigger.click=!i._activeTrigger.click,i._isWithActiveTrigger()?i._enter(null,i):i._leave(null,i)}else{if(e(this.getTipElement()).hasClass(He))return void this._leave(null,this)
this._enter(null,this)}},n.dispose=function(){clearTimeout(this._timeout),e.removeData(this.element,this.constructor.DATA_KEY),e(this.element).off(this.constructor.EVENT_KEY),e(this.element).closest(".modal").off("hide.bs.modal"),this.tip&&e(this.tip).remove(),this._isEnabled=null,this._timeout=null,this._hoverState=null,(this._activeTrigger=null)!==this._popper&&this._popper.destroy(),this._popper=null,this.element=null,this.config=null,this.tip=null},n.show=function(){var t=this
if("none"===e(this.element).css("display"))throw new Error("Please use show on visible elements")
var n=e.Event(this.constructor.Event.SHOW)
if(this.isWithContent()&&this._isEnabled){e(this.element).trigger(n)
var i=s.findShadowRoot(this.element),o=e.contains(null!==i?i:this.element.ownerDocument.documentElement,this.element)
if(n.isDefaultPrevented()||!o)return
var r=this.getTipElement(),a=s.getUID(this.constructor.NAME)
r.setAttribute("id",a),this.element.setAttribute("aria-describedby",a),this.setContent(),this.config.animation&&e(r).addClass(Pe)
var l="function"==typeof this.config.placement?this.config.placement.call(this,r,this.element):this.config.placement,c=this._getAttachment(l)
this.addAttachmentClass(c)
var h=this._getContainer()
e(r).data(this.constructor.DATA_KEY,this),e.contains(this.element.ownerDocument.documentElement,this.tip)||e(r).appendTo(h),e(this.element).trigger(this.constructor.Event.INSERTED),this._popper=new Bt(this.element,r,{placement:c,modifiers:{offset:this._getOffset(),flip:{behavior:this.config.fallbackPlacement},arrow:{element:".arrow"},preventOverflow:{boundariesElement:this.config.boundary}},onCreate:function(e){e.originalPlacement!==e.placement&&t._handlePopperPlacementChange(e)},onUpdate:function(e){return t._handlePopperPlacementChange(e)}}),e(r).addClass(He),"ontouchstart"in document.documentElement&&e(document.body).children().on("mouseover",null,e.noop)
var u=function(){t.config.animation&&t._fixTransition()
var n=t._hoverState
t._hoverState=null,e(t.element).trigger(t.constructor.Event.SHOWN),"out"===n&&t._leave(null,t)}
if(e(this.tip).hasClass(Pe)){var f=s.getTransitionDurationFromElement(this.tip)
e(this.tip).one(s.TRANSITION_END,u).emulateTransitionEnd(f)}else u()}},n.hide=function(t){var n=this,i=this.getTipElement(),o=e.Event(this.constructor.Event.HIDE),r=function(){n._hoverState!==Le&&i.parentNode&&i.parentNode.removeChild(i),n._cleanTipClass(),n.element.removeAttribute("aria-describedby"),e(n.element).trigger(n.constructor.Event.HIDDEN),null!==n._popper&&n._popper.destroy(),t&&t()}
if(e(this.element).trigger(o),!o.isDefaultPrevented()){if(e(i).removeClass(He),"ontouchstart"in document.documentElement&&e(document.body).children().off("mouseover",null,e.noop),this._activeTrigger.click=!1,this._activeTrigger[Re]=!1,this._activeTrigger[je]=!1,e(this.tip).hasClass(Pe)){var a=s.getTransitionDurationFromElement(i)
e(i).one(s.TRANSITION_END,r).emulateTransitionEnd(a)}else r()
this._hoverState=""}},n.update=function(){null!==this._popper&&this._popper.scheduleUpdate()},n.isWithContent=function(){return Boolean(this.getTitle())},n.addAttachmentClass=function(t){e(this.getTipElement()).addClass(De+"-"+t)},n.getTipElement=function(){return this.tip=this.tip||e(this.config.template)[0],this.tip},n.setContent=function(){var t=this.getTipElement()
this.setElementContent(e(t.querySelectorAll(".tooltip-inner")),this.getTitle()),e(t).removeClass(Pe+" "+He)},n.setElementContent=function(t,n){"object"!=typeof n||!n.nodeType&&!n.jquery?this.config.html?(this.config.sanitize&&(n=be(n,this.config.whiteList,this.config.sanitizeFn)),t.html(n)):t.text(n):this.config.html?e(n).parent().is(t)||t.empty().append(n):t.text(e(n).text())},n.getTitle=function(){var t=this.element.getAttribute("data-original-title")
return t||(t="function"==typeof this.config.title?this.config.title.call(this.element):this.config.title),t},n._getOffset=function(){var t=this,e={}
return"function"==typeof this.config.offset?e.fn=function(e){return e.offsets=o({},e.offsets,t.config.offset(e.offsets,t.element)||{}),e}:e.offset=this.config.offset,e},n._getContainer=function(){return!1===this.config.container?document.body:s.isElement(this.config.container)?e(this.config.container):e(document).find(this.config.container)},n._getAttachment=function(t){return Ne[t.toUpperCase()]},n._setListeners=function(){var t=this
this.config.trigger.split(" ").forEach(function(n){if("click"===n)e(t.element).on(t.constructor.Event.CLICK,t.config.selector,function(e){return t.toggle(e)})
else if("manual"!==n){var i=n===je?t.constructor.Event.MOUSEENTER:t.constructor.Event.FOCUSIN,o=n===je?t.constructor.Event.MOUSELEAVE:t.constructor.Event.FOCUSOUT
e(t.element).on(i,t.config.selector,function(e){return t._enter(e)}).on(o,t.config.selector,function(e){return t._leave(e)})}}),e(this.element).closest(".modal").on("hide.bs.modal",function(){t.element&&t.hide()}),this.config.selector?this.config=o({},this.config,{trigger:"manual",selector:""}):this._fixTitle()},n._fixTitle=function(){var t=typeof this.element.getAttribute("data-original-title");(this.element.getAttribute("title")||"string"!==t)&&(this.element.setAttribute("data-original-title",this.element.getAttribute("title")||""),this.element.setAttribute("title",""))},n._enter=function(t,n){var i=this.constructor.DATA_KEY;(n=n||e(t.currentTarget).data(i))||(n=new this.constructor(t.currentTarget,this._getDelegateConfig()),e(t.currentTarget).data(i,n)),t&&(n._activeTrigger["focusin"===t.type?Re:je]=!0),e(n.getTipElement()).hasClass(He)||n._hoverState===Le?n._hoverState=Le:(clearTimeout(n._timeout),n._hoverState=Le,n.config.delay&&n.config.delay.show?n._timeout=setTimeout(function(){n._hoverState===Le&&n.show()},n.config.delay.show):n.show())},n._leave=function(t,n){var i=this.constructor.DATA_KEY;(n=n||e(t.currentTarget).data(i))||(n=new this.constructor(t.currentTarget,this._getDelegateConfig()),e(t.currentTarget).data(i,n)),t&&(n._activeTrigger["focusout"===t.type?Re:je]=!1),n._isWithActiveTrigger()||(clearTimeout(n._timeout),n._hoverState="out",n.config.delay&&n.config.delay.hide?n._timeout=setTimeout(function(){"out"===n._hoverState&&n.hide()},n.config.delay.hide):n.hide())},n._isWithActiveTrigger=function(){for(var t in this._activeTrigger)if(this._activeTrigger[t])return!0
return!1},n._getConfig=function(t){var n=e(this.element).data()
return Object.keys(n).forEach(function(t){-1!==Ae.indexOf(t)&&delete n[t]}),"number"==typeof(t=o({},this.constructor.Default,n,"object"==typeof t&&t?t:{})).delay&&(t.delay={show:t.delay,hide:t.delay}),"number"==typeof t.title&&(t.title=t.title.toString()),"number"==typeof t.content&&(t.content=t.content.toString()),s.typeCheckConfig(we,t,this.constructor.DefaultType),t.sanitize&&(t.template=be(t.template,t.whiteList,t.sanitizeFn)),t},n._getDelegateConfig=function(){var t={}
if(this.config)for(var e in this.config)this.constructor.Default[e]!==this.config[e]&&(t[e]=this.config[e])
return t},n._cleanTipClass=function(){var t=e(this.getTipElement()),n=t.attr("class").match(Ie)
null!==n&&n.length&&t.removeClass(n.join(""))},n._handlePopperPlacementChange=function(t){var e=t.instance
this.tip=e.popper,this._cleanTipClass(),this.addAttachmentClass(this._getAttachment(t.placement))},n._fixTransition=function(){var t=this.getTipElement(),n=this.config.animation
null===t.getAttribute("x-placement")&&(e(t).removeClass(Pe),this.config.animation=!1,this.hide(),this.show(),this.config.animation=n)},t._jQueryInterface=function(n){return this.each(function(){var i=e(this).data(Ce),o="object"==typeof n&&n
if((i||!/dispose|hide/.test(n))&&(i||(i=new t(this,o),e(this).data(Ce,i)),"string"==typeof n)){if(void 0===i[n])throw new TypeError('No method named "'+n+'"')
i[n]()}})},i(t,null,[{key:"VERSION",get:function(){return"4.3.1"}},{key:"Default",get:function(){return ke}},{key:"NAME",get:function(){return we}},{key:"DATA_KEY",get:function(){return Ce}},{key:"Event",get:function(){return xe}},{key:"EVENT_KEY",get:function(){return Te}},{key:"DefaultType",get:function(){return Oe}}]),t}()
e.fn[we]=Fe._jQueryInterface,e.fn[we].Constructor=Fe,e.fn[we].noConflict=function(){return e.fn[we]=Se,Fe._jQueryInterface}
var Me="popover",We="bs.popover",Ue="."+We,Be=e.fn[Me],qe="bs-popover",Ke=new RegExp("(^|\\s)"+qe+"\\S+","g"),Qe=o({},Fe.Default,{placement:"right",trigger:"click",content:"",template:'<div class="popover" role="tooltip"><div class="arrow"></div><h3 class="popover-header"></h3><div class="popover-body"></div></div>'}),Ve=o({},Fe.DefaultType,{content:"(string|element|function)"}),Ye={HIDE:"hide"+Ue,HIDDEN:"hidden"+Ue,SHOW:"show"+Ue,SHOWN:"shown"+Ue,INSERTED:"inserted"+Ue,CLICK:"click"+Ue,FOCUSIN:"focusin"+Ue,FOCUSOUT:"focusout"+Ue,MOUSEENTER:"mouseenter"+Ue,MOUSELEAVE:"mouseleave"+Ue},ze=function(t){var n,o
function r(){return t.apply(this,arguments)||this}o=t,(n=r).prototype=Object.create(o.prototype),(n.prototype.constructor=n).__proto__=o
var s=r.prototype
return s.isWithContent=function(){return this.getTitle()||this._getContent()},s.addAttachmentClass=function(t){e(this.getTipElement()).addClass(qe+"-"+t)},s.getTipElement=function(){return this.tip=this.tip||e(this.config.template)[0],this.tip},s.setContent=function(){var t=e(this.getTipElement())
this.setElementContent(t.find(".popover-header"),this.getTitle())
var n=this._getContent()
"function"==typeof n&&(n=n.call(this.element)),this.setElementContent(t.find(".popover-body"),n),t.removeClass("fade show")},s._getContent=function(){return this.element.getAttribute("data-content")||this.config.content},s._cleanTipClass=function(){var t=e(this.getTipElement()),n=t.attr("class").match(Ke)
null!==n&&0<n.length&&t.removeClass(n.join(""))},r._jQueryInterface=function(t){return this.each(function(){var n=e(this).data(We),i="object"==typeof t?t:null
if((n||!/dispose|hide/.test(t))&&(n||(n=new r(this,i),e(this).data(We,n)),"string"==typeof t)){if(void 0===n[t])throw new TypeError('No method named "'+t+'"')
n[t]()}})},i(r,null,[{key:"VERSION",get:function(){return"4.3.1"}},{key:"Default",get:function(){return Qe}},{key:"NAME",get:function(){return Me}},{key:"DATA_KEY",get:function(){return We}},{key:"Event",get:function(){return Ye}},{key:"EVENT_KEY",get:function(){return Ue}},{key:"DefaultType",get:function(){return Ve}}]),r}(Fe)
e.fn[Me]=ze._jQueryInterface,e.fn[Me].Constructor=ze,e.fn[Me].noConflict=function(){return e.fn[Me]=Be,ze._jQueryInterface}
var Xe="scrollspy",Ge="bs.scrollspy",$e="."+Ge,Je=e.fn[Xe],Ze={offset:10,method:"auto",target:""},tn={offset:"number",method:"string",target:"(string|element)"},en={ACTIVATE:"activate"+$e,SCROLL:"scroll"+$e,LOAD_DATA_API:"load"+$e+".data-api"},nn="active",on=".nav, .list-group",rn=".nav-link",sn=".list-group-item",an=".dropdown-item",ln="position",cn=function(){function t(t,n){var i=this
this._element=t,this._scrollElement="BODY"===t.tagName?window:t,this._config=this._getConfig(n),this._selector=this._config.target+" "+rn+","+this._config.target+" "+sn+","+this._config.target+" "+an,this._offsets=[],this._targets=[],this._activeTarget=null,this._scrollHeight=0,e(this._scrollElement).on(en.SCROLL,function(t){return i._process(t)}),this.refresh(),this._process()}var n=t.prototype
return n.refresh=function(){var t=this,n=this._scrollElement===this._scrollElement.window?"offset":ln,i="auto"===this._config.method?n:this._config.method,o=i===ln?this._getScrollTop():0
this._offsets=[],this._targets=[],this._scrollHeight=this._getScrollHeight(),[].slice.call(document.querySelectorAll(this._selector)).map(function(t){var n,r=s.getSelectorFromElement(t)
if(r&&(n=document.querySelector(r)),n){var a=n.getBoundingClientRect()
if(a.width||a.height)return[e(n)[i]().top+o,r]}return null}).filter(function(t){return t}).sort(function(t,e){return t[0]-e[0]}).forEach(function(e){t._offsets.push(e[0]),t._targets.push(e[1])})},n.dispose=function(){e.removeData(this._element,Ge),e(this._scrollElement).off($e),this._element=null,this._scrollElement=null,this._config=null,this._selector=null,this._offsets=null,this._targets=null,this._activeTarget=null,this._scrollHeight=null},n._getConfig=function(t){if("string"!=typeof(t=o({},Ze,"object"==typeof t&&t?t:{})).target){var n=e(t.target).attr("id")
n||(n=s.getUID(Xe),e(t.target).attr("id",n)),t.target="#"+n}return s.typeCheckConfig(Xe,t,tn),t},n._getScrollTop=function(){return this._scrollElement===window?this._scrollElement.pageYOffset:this._scrollElement.scrollTop},n._getScrollHeight=function(){return this._scrollElement.scrollHeight||Math.max(document.body.scrollHeight,document.documentElement.scrollHeight)},n._getOffsetHeight=function(){return this._scrollElement===window?window.innerHeight:this._scrollElement.getBoundingClientRect().height},n._process=function(){var t=this._getScrollTop()+this._config.offset,e=this._getScrollHeight(),n=this._config.offset+e-this._getOffsetHeight()
if(this._scrollHeight!==e&&this.refresh(),n<=t){var i=this._targets[this._targets.length-1]
this._activeTarget!==i&&this._activate(i)}else{if(this._activeTarget&&t<this._offsets[0]&&0<this._offsets[0])return this._activeTarget=null,void this._clear()
for(var o=this._offsets.length;o--;)this._activeTarget!==this._targets[o]&&t>=this._offsets[o]&&(void 0===this._offsets[o+1]||t<this._offsets[o+1])&&this._activate(this._targets[o])}},n._activate=function(t){this._activeTarget=t,this._clear()
var n=this._selector.split(",").map(function(e){return e+'[data-target="'+t+'"],'+e+'[href="'+t+'"]'}),i=e([].slice.call(document.querySelectorAll(n.join(","))))
i.hasClass("dropdown-item")?(i.closest(".dropdown").find(".dropdown-toggle").addClass(nn),i.addClass(nn)):(i.addClass(nn),i.parents(on).prev(rn+", "+sn).addClass(nn),i.parents(on).prev(".nav-item").children(rn).addClass(nn)),e(this._scrollElement).trigger(en.ACTIVATE,{relatedTarget:t})},n._clear=function(){[].slice.call(document.querySelectorAll(this._selector)).filter(function(t){return t.classList.contains(nn)}).forEach(function(t){return t.classList.remove(nn)})},t._jQueryInterface=function(n){return this.each(function(){var i=e(this).data(Ge)
if(i||(i=new t(this,"object"==typeof n&&n),e(this).data(Ge,i)),"string"==typeof n){if(void 0===i[n])throw new TypeError('No method named "'+n+'"')
i[n]()}})},i(t,null,[{key:"VERSION",get:function(){return"4.3.1"}},{key:"Default",get:function(){return Ze}}]),t}()
e(window).on(en.LOAD_DATA_API,function(){for(var t=[].slice.call(document.querySelectorAll('[data-spy="scroll"]')),n=t.length;n--;){var i=e(t[n])
cn._jQueryInterface.call(i,i.data())}}),e.fn[Xe]=cn._jQueryInterface,e.fn[Xe].Constructor=cn,e.fn[Xe].noConflict=function(){return e.fn[Xe]=Je,cn._jQueryInterface}
var hn="bs.tab",un="."+hn,fn=e.fn.tab,dn={HIDE:"hide"+un,HIDDEN:"hidden"+un,SHOW:"show"+un,SHOWN:"shown"+un,CLICK_DATA_API:"click"+un+".data-api"},pn="active",mn=".active",gn="> li > .active",_n=function(){function t(t){this._element=t}var n=t.prototype
return n.show=function(){var t=this
if(!(this._element.parentNode&&this._element.parentNode.nodeType===Node.ELEMENT_NODE&&e(this._element).hasClass(pn)||e(this._element).hasClass("disabled"))){var n,i,o=e(this._element).closest(".nav, .list-group")[0],r=s.getSelectorFromElement(this._element)
if(o){var a="UL"===o.nodeName||"OL"===o.nodeName?gn:mn
i=(i=e.makeArray(e(o).find(a)))[i.length-1]}var l=e.Event(dn.HIDE,{relatedTarget:this._element}),c=e.Event(dn.SHOW,{relatedTarget:i})
if(i&&e(i).trigger(l),e(this._element).trigger(c),!c.isDefaultPrevented()&&!l.isDefaultPrevented()){r&&(n=document.querySelector(r)),this._activate(this._element,o)
var h=function(){var n=e.Event(dn.HIDDEN,{relatedTarget:t._element}),o=e.Event(dn.SHOWN,{relatedTarget:i})
e(i).trigger(n),e(t._element).trigger(o)}
n?this._activate(n,n.parentNode,h):h()}}},n.dispose=function(){e.removeData(this._element,hn),this._element=null},n._activate=function(t,n,i){var o=this,r=(!n||"UL"!==n.nodeName&&"OL"!==n.nodeName?e(n).children(mn):e(n).find(gn))[0],a=i&&r&&e(r).hasClass("fade"),l=function(){return o._transitionComplete(t,r,i)}
if(r&&a){var c=s.getTransitionDurationFromElement(r)
e(r).removeClass("show").one(s.TRANSITION_END,l).emulateTransitionEnd(c)}else l()},n._transitionComplete=function(t,n,i){if(n){e(n).removeClass(pn)
var o=e(n.parentNode).find("> .dropdown-menu .active")[0]
o&&e(o).removeClass(pn),"tab"===n.getAttribute("role")&&n.setAttribute("aria-selected",!1)}if(e(t).addClass(pn),"tab"===t.getAttribute("role")&&t.setAttribute("aria-selected",!0),s.reflow(t),t.classList.contains("fade")&&t.classList.add("show"),t.parentNode&&e(t.parentNode).hasClass("dropdown-menu")){var r=e(t).closest(".dropdown")[0]
if(r){var a=[].slice.call(r.querySelectorAll(".dropdown-toggle"))
e(a).addClass(pn)}t.setAttribute("aria-expanded",!0)}i&&i()},t._jQueryInterface=function(n){return this.each(function(){var i=e(this),o=i.data(hn)
if(o||(o=new t(this),i.data(hn,o)),"string"==typeof n){if(void 0===o[n])throw new TypeError('No method named "'+n+'"')
o[n]()}})},i(t,null,[{key:"VERSION",get:function(){return"4.3.1"}}]),t}()
e(document).on(dn.CLICK_DATA_API,'[data-toggle="tab"], [data-toggle="pill"], [data-toggle="list"]',function(t){t.preventDefault(),_n._jQueryInterface.call(e(this),"show")}),e.fn.tab=_n._jQueryInterface,e.fn.tab.Constructor=_n,e.fn.tab.noConflict=function(){return e.fn.tab=fn,_n._jQueryInterface}
var vn="toast",yn="bs.toast",En="."+yn,bn=e.fn[vn],wn={CLICK_DISMISS:"click.dismiss"+En,HIDE:"hide"+En,HIDDEN:"hidden"+En,SHOW:"show"+En,SHOWN:"shown"+En},Cn="show",Tn="showing",Sn={animation:"boolean",autohide:"boolean",delay:"number"},Dn={animation:!0,autohide:!0,delay:500},In=function(){function t(t,e){this._element=t,this._config=this._getConfig(e),this._timeout=null,this._setListeners()}var n=t.prototype
return n.show=function(){var t=this
e(this._element).trigger(wn.SHOW),this._config.animation&&this._element.classList.add("fade")
var n=function(){t._element.classList.remove(Tn),t._element.classList.add(Cn),e(t._element).trigger(wn.SHOWN),t._config.autohide&&t.hide()}
if(this._element.classList.remove("hide"),this._element.classList.add(Tn),this._config.animation){var i=s.getTransitionDurationFromElement(this._element)
e(this._element).one(s.TRANSITION_END,n).emulateTransitionEnd(i)}else n()},n.hide=function(t){var n=this
this._element.classList.contains(Cn)&&(e(this._element).trigger(wn.HIDE),t?this._close():this._timeout=setTimeout(function(){n._close()},this._config.delay))},n.dispose=function(){clearTimeout(this._timeout),this._timeout=null,this._element.classList.contains(Cn)&&this._element.classList.remove(Cn),e(this._element).off(wn.CLICK_DISMISS),e.removeData(this._element,yn),this._element=null,this._config=null},n._getConfig=function(t){return t=o({},Dn,e(this._element).data(),"object"==typeof t&&t?t:{}),s.typeCheckConfig(vn,t,this.constructor.DefaultType),t},n._setListeners=function(){var t=this
e(this._element).on(wn.CLICK_DISMISS,'[data-dismiss="toast"]',function(){return t.hide(!0)})},n._close=function(){var t=this,n=function(){t._element.classList.add("hide"),e(t._element).trigger(wn.HIDDEN)}
if(this._element.classList.remove(Cn),this._config.animation){var i=s.getTransitionDurationFromElement(this._element)
e(this._element).one(s.TRANSITION_END,n).emulateTransitionEnd(i)}else n()},t._jQueryInterface=function(n){return this.each(function(){var i=e(this),o=i.data(yn)
if(o||(o=new t(this,"object"==typeof n&&n),i.data(yn,o)),"string"==typeof n){if(void 0===o[n])throw new TypeError('No method named "'+n+'"')
o[n](this)}})},i(t,null,[{key:"VERSION",get:function(){return"4.3.1"}},{key:"DefaultType",get:function(){return Sn}},{key:"Default",get:function(){return Dn}}]),t}()
e.fn[vn]=In._jQueryInterface,e.fn[vn].Constructor=In,e.fn[vn].noConflict=function(){return e.fn[vn]=bn,In._jQueryInterface},function(){if(void 0===e)throw new TypeError("Bootstrap's JavaScript requires jQuery. jQuery must be included before Bootstrap's JavaScript.")
var t=e.fn.jquery.split(" ")[0].split(".")
if(t[0]<2&&t[1]<9||1===t[0]&&9===t[1]&&t[2]<1||4<=t[0])throw new Error("Bootstrap's JavaScript requires at least jQuery v1.9.1 but less than v4.0.0")}(),t.Util=s,t.Alert=f,t.Button=w,t.Carousel=R,t.Collapse=G,t.Dropdown=ie,t.Modal=_e,t.Popover=ze,t.Scrollspy=cn,t.Tab=_n,t.Toast=In,t.Tooltip=Fe,Object.defineProperty(t,"__esModule",{value:!0})})
