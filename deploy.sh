# ~/bin/bash
systemctl stop yig && cp yig /usr/bin/yig && cp plugins/* /etc/yig/plugins && systemctl start yig
