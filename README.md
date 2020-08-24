### X-MAN
X-man is a packet capture tool for http/https requests, which can be used to record, 
inject, reproduce, export, etc. of network requests from mobile devices/pcs. 

### Install
```
go get github.com/clearcodecn/x-man
```

### Usage: 
1. First you should install and trust root certificate. you can download it from `http://127.0.0.1/-/cert`
2. Setup your http proxy server to `http://127.0.0.1:3344`
3. start request. you can open ui console `http://127.0.0.1:3344/-/` to see the details.

### Fetures
- [x] javascript(text) inject
- [x] request replay
- [x] export request/response detail
- [ ] using custom certificate 
- [ ] body replace
- [ ] ui change 