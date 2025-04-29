#!/bin/sh
#
# suricata-reload.sh
# 1. 动态提取 WAF 反代端口，更新 Suricata 的 BPF 过滤表达式
# 2. 触发已运行 Suricata 进程重读最新配置（热重载）

# —— 第 1 步：从 HAProxy 配置中提取所有 bind 端口 —— 
# 假设你的 HAProxy 配置挂载在 /etc/haproxy/haproxy.cfg
ports=$(grep -E '^\s*bind\s+' /etc/haproxy/haproxy.cfg \
  | grep -oE ':[0-9]+' | tr -d ':' | sort -u)

# —— 第 2 步：拼接成 BPF 过滤表达式 —— 
# 例如：tcp port 80 or tcp port 8080 or tcp port 8443
bpf=""
IFS=','; for p in $ports; do
  if [ -n "$bpf" ]; then
    bpf="$bpf or "
  fi
  bpf="${bpf}tcp port ${p}"
done
unset IFS

# —— 第 3 步：更新 suricata.yaml 中 bpf-filter 条目 —— 
# 假设你在 suricata.yaml 中预留了一行：
#     bpf-filter: ""
# 下面这条 sed 会把它替换成实际的过滤串
sed -i "s#^\(\s*bpf-filter:\).*#\1 \"${bpf}\"#g" /etc/suricata/suricata.yaml

# —— 第 4 步：触发 Suricata 热重载 —— 
# 如果 suricata.yaml 中开启了 unix-command
#    unix-command:
#      enabled: yes
#      filename: /var/run/suricata-command.socket
# 则可以通过 socket 下发 reload-config
if [ -S /var/run/suricata-command.socket ]; then
  echo 'reload-config' | socat - UNIX-CONNECT:/var/run/suricata-command.socket
else
  # 否则退回到发送 HUP 信号
  pid=$(pidof suricata)
  [ -n "$
