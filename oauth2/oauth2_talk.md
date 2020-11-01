## 白话OAuth2.0

### OAuth2.0 是什么和解决什么问题

简单一句话，OAuth2.0 是一种授权协议。比如，我们使用微信的账号登录某个第三方app, 其实是微信给第三方应用授权了，第三方应用就能拿到用户微信账号的一些基本信息，包括头像，昵称等。

那么，为什么非要用OAuth2.0的方式进行授权呢，为什么不用其他的方式呢？比如直接给第三方应用微信账号的用户名和密码，第三方应用拿到这个用户名和密码从微信哪里获取信息。这种方式显然是不可行的。首先，这涉及到信息泄露，你给了第三方应用微信的用户名和密码，第三方应用可能有意或无意的泄露了你的微信密码。其次，你只想第三方应用访问你的微信的头像，昵称等基本信息，如果把密码给了第三方应用，其他信息第三方应用可能会拿去。当然，这里最大的问题是，你不信任第三方应用，我不会把第三方密码给它。 通过使用OAuth2.0的授权机制，我们给第三方应用一个临时token, 第三方应用使用这个临时token从微信那里获得有限的信息。

那总结来说，OAuth 2.0 这种授权协议，就是保证第三方应用只有在获得授权之后，才可以进一步访问授权者的数据。

### OAuth2.0完整的授权流程

以微信登录为例

![oauth2-wechat-login](https://github.com/checkking/notes/blob/master/imgs/oauth_img_1.png)


1. 第三方发起微信授权登录请求，微信用户允许授权第三方应用后，微信会拉起应用或重定向到第三方网站，并且带上授权临时票据code参数；

2. 通过code参数加上AppID和AppSecret等，通过API换取access_token；

3. 通过access_token进行接口调用，获取用户基本数据资源或帮助用户实现基本操作。

其实，第三方应用想要通过微信的账号进行登录的前提是，第三方应用要在微信的开放平台进行登记，登记的时候需要登记重定向url, scope等信息，用于开放平台的验证和跳转回第三方应用的页面用。其中，AppID,AppSecret是开放平台给第三方应用的凭证。上面流程中的第2步，为了不暴露appsecret, 是需要后端来调用微信开放平台接口的。

#### 上面流程为什么需要临时票据(code)

有两个原因，第一，访问令牌(token) 要求有极高的安全保密性, 不能暴露在浏览器上，需要通过后端服务来获取，以最大限度保证访问令牌的安全性。第二，通过临时票据的方式进行重定向到第三方应用页面，在保证安全的前提下，能够在给用户授权之后，跳回用户的操作页面，提升了用户体验。

### OAuth2.0 四种授权模式
#### 授权码模式(authorization code)
最完备，最安全的授权模式。特点是客户端先从授权服务器拿到授权码(code)，经过后端服务器用授权码从授权服务器换取token, 具体流程就是上面讲的例子。

#### 隐式模式(implicit)

适用于单页应用，没有后台的应用（比如纯JavaScript）。特点是token直接返回给浏览器(User Agent)，不经过第三方应用后台。具体流程如下:

![oauth2_img_2.png](https://github.com/checkking/notes/blob/master/imgs/oauth2_img_2.png)

#### 资源拥有者凭证模式(resource owner password credential)
用户向客户端提供密码，客户端用密码向资源服务器获取token。客户端不得存储用户名和密码，适用于高度信任的情况。这种情况比较适合第三方应用和资源服务器之间是高度信任的情况。

![oauth2_img_3.png](https://github.com/checkking/notes/blob/master/imgs/oauth2_img_3.png)

#### 客户端凭证模式(client credentials)

客户端模式，适合服务之间调用的授权，没有前端参与。比如A应用通过client_id和client_secret向B应用获取资源的方式。
![oauth2_img_4.png](https://github.com/checkking/notes/blob/master/imgs/oauth2_img_4.png)

### 刷新令牌

在授权码模式中，如果access token过期了，需要重新授权获取access token, 这就需要用户再次手动授权，用户体验不好。这里引入刷新令牌，来避免用户反复手动授权。

refresh token是和access token一起颁发给第三方应用的，当第三方应用发现access token过期了，便用refresh token请求授权服务器获取access token和新的refresh token，第三方应用保存新的refresh token。

### JWT (JSON Web Token)

OAuth2.0并没有约束access token的生成规则，只要符合唯一性，不连续性和不可猜测性就可以。 授权服务器颁发给第三方应用的access token, 第三方应用根据access token从资源服务器获取数据，资源服务器是需要校验access token的合法性的，而校验就是需要通过rpc 调用授权服务器。在现在微服务这种架构下，所有受保护的模块都需要请求授权服务器，这种方式对授权服务器的压力就太大了。而JWT这种令牌具有自解释性，自身包含了从授权服务器获取的一些信息。

#### JWT的格式

JWT结构化体包含HEADER, PAYLOAD和SIGNATURE三部分，HEADER表示JWT头部，一般包含类型和加密算法信息。PAYLOAD表示JWT的数据体，用来承载一些token信息数据。SIGNATURE 表示JWT的信息签名，避免jwt token在网络传输过程中被人篡改和数据信息泄露。生成jwt的过程还需要做Base64处理，以避免乱码问题。下面是一个jwt token的例子:
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IlhpYW9taW5nIiwiaWF0IjoxNTE2MjM5MDIyfQ.q25mDFCAlWlqR3YMJr2Jds3ntBTyGItdvijiEtuJJ2E
```

我们可以到jwt.io上去解码查看具体的内容如下:

![oauth2_img_5.png](https://github.com/checkking/notes/blob/master/imgs/oauth2_img5.png)

#### JWT的不足

jwt的有点很明显，就是jwt token的自解释性，将token校验放到了各个资源业务模块，减少了对授权服务器的压力。但是如果想吊销一个有问题的token, 就比较难了，因为jwt token由授权服务器颁发之后，就不由授权服务器控制了。

为了解决吊销的问题，授权服务器还是需要能够控制token，我们可以做一个优化，颁发token的时候记录，返回普通的access token， 然后在客户端应用通过网关调用资源服务器的时候，用access token换取jwt token，在内网中利用jwt token调接口读取数据。

### 微服务安全架构

![oauth2_img_6.png](https://github.com/checkking/notes/blob/master/imgs/oauth2_img6.png)

1. 客户应用先去授权服务器拿到access token (授权服务器颁发access token，并做jwt映射，缓存在redis中)
2. 客户端携带access token访问网关，网关拿access token去授权服务器校验并换取jwt token(优化：直接读取redis)
3. 网关将请求携带jwt token路由到下游各微服务.
4. 由于jwt的自解释性，下游各模块能校验请求。

### 一个Demo

实现了一个很简单的demo, 包含最基本的oauth2流程。github地址: https://github.com/checkking/oauth2_practice
