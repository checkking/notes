### OAuth2

#### OAuth2解决的问题
OAuth2要解决的问题是，第三方应用想访问用户存储在资源服务器上的资源需要授权的问题。简单说，OAuth2是一个授权机制。数据的所有者告诉系统，同意第三方应用访问系统数据。系统从而产生一个短期的令牌(token)，供第三方应用访问系统数据。

#### OAuth2各个角色
* 资源所有者, 也就是用户，在参与OAuth2的流程中主要作用是触发授权动作的主体.
* 第三方应用， 也就是资源所有者需要授权访问资源服务器的对象.
* 资源服务器, 存放用户的资源
* 授权服务器， 用来颁发授权码和授权令牌(token)

#### 4种授权模式
* 授权码模式(authorization code) 最完备，最安全的授权模式. 特点是客户端先从授权服务器拿到授权码，经过后端服务器用授权码从授权服务器换取token。
* 简化模式(implicit) 适用于单页应用，没有后台的应用。特点是token直接返回给浏览器(User Agent)，不经过第三方应用后台。
* 密码模式(resource owner password credentials). 用户向客户端提供密码，客户端用密码向资源服务器获取token。客户端不得存储用户名和密码，适用于高度信任的情况。
* 客户端模式（client credentials）

#### 授权码模式

![oauth2-code1](https://github.com/checkking/notes/blob/master/imgs/oath_code1.png)

步骤：
1）用户访问第三方应用(客户端), 后者将用户重定向到授权服务器认证页面
```
GET /authorize?response_type=code&client_id=s6BhdRkqt3&state=xyz
        &redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb HTTP/1.1
Host: server.example.com
```
2) 用户选择是否给与客户端授权。
3) 假设用户给与授权，授权服务器将用户重定向到约定好的"重定向URI"，并同时附上一个授权码.

```
HTTP/1.1 302 Found
Location: https://client.example.com/cb?code=SplxlOBeZQQYbYS6WxSbIA
          &state=xyz
```

3) 客户端收到授权码，和事先约定好的"重定向URI"一起向授权服务器发起获取token的请求。这一步是在客户端的后台的服务器上完成的，对用户不可见。

```
POST /token HTTP/1.1
Host: server.example.com
Authorization: Basic czZCaGRSa3F0MzpnWDFmQmF0M2JW
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code&code=SplxlOBeZQQYbYS6WxSbIA
&redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb
```

4) 授权服务器对授权码和重定向URI进行核对，确认无误后，向客户端颁发访问令牌（access token）和更新令牌（refresh token)。

```
     HTTP/1.1 200 OK
     Content-Type: application/json;charset=UTF-8
     Cache-Control: no-store
     Pragma: no-cache

     {
       "access_token":"2YotnFZFEjr1zCsicMWpAA",
       "token_type":"example",
       "expires_in":3600,
       "refresh_token":"tGzv3JOkF0XG5Qx2TlKWIA",
       "example_parameter":"example_value"
     }

```

#####  为什么一定要有授权码

* 访问令牌(token) 要求有极高的安全保密性, 不能暴露在浏览器上，需要通过后端服务来获取，以最大限度保证访问令牌的安全性。
* 通过授权码的方式，在保证安全的前提下，能够在给用户授权之后，跳回用户的操作页面，提升了用户体验.

### 微服务安全架构

![oauth2-arch](https://github.com/checkking/notes/blob/master/imgs/arch_oauth2.png)

1. 客户应用先去授权服务器拿到access token (授权服务器颁发access token，并做jwt映射，缓存在redis中)
2. 客户端携带access token访问网关，网关拿access token去授权服务器校验并换取jwt token(优化：直接读取redis)
3. 网关将请求携带jwt token路由到下游各微服务.
4. 由于jwt的自解释性，下游各模块能校验请求。
