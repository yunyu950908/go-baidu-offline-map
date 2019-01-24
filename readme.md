#### 使用方法

- step 0.
拉代码, 装依赖
```bash
git clone git@github.com:yunyu950908/go-baidu-offline-map.git --depth 1

cd go-baidu-offline-map

# 如果你不需要预览，可以不装该依赖，略过这一步
npm install --registry=https://registry.npm.taobao.org/
```

- step 1.
根据[百度拾取坐标系统](http://api.map.baidu.com/lbsapi/getpoint/index.html)查找要下载区域的经纬度信息。
![image](https://user-images.githubusercontent.com/25625252/51693014-0e7fec00-2039-11e9-966d-03975e95b496.png)

- step 2.
按提示输入数据, 开始下载
![image](https://user-images.githubusercontent.com/25625252/51693502-0f654d80-203a-11e9-9fd4-662d0cd7b56f.png)

- step 3.
预览下载的区域地图
下载完成后会在存储贴片的目录生成一个 index.html, 使用浏览器打开即可预览

