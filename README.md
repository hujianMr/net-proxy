## go-huj-net-proxy 网络代理服务
### 背景
&nbsp;&nbsp;&nbsp;&nbsp; 让家里的linux服务器的端口能让外网访问

### 实现方式
&nbsp;&nbsp;&nbsp;&nbsp; 1. 开发服务端，服务端部署在一个配置低的阿里云服务器    
&nbsp;&nbsp;&nbsp;&nbsp; 2. 开发客户端服务，部署在家里的linux服务器  
&nbsp;&nbsp;&nbsp;&nbsp; 3. 客户端跟服务端通讯，上传需要映射的端口，跟服务端建立长连接   
&nbsp;&nbsp;&nbsp;&nbsp; 4. 服务端接收客户端需要映射的端口开启不同的端口tpc监听 
&nbsp;&nbsp;&nbsp;&nbsp; 5. 在阿里云控制台 安全规则 那里开启对应客户端需要映射的端口   
&nbsp;&nbsp;&nbsp;&nbsp; 6. 每当一次远程请求过来，根据监听端口在map当中找到客户端服务服务长连接的哪个conn, 以此进行转发通讯    