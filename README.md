# go-proxymanager
a command line utility written in go to manage reverse proxies on an nginx server. Also acts as a frontend/traffic manager for onpremisis kubernetes clusters.

## features
proxymanager lb
    proxymanager lb new <cluster>
    proxymanager lb remove <cluster>
    proxymanager lb list
    proxymanager lb <cluster> status
    proxymanager lb <cluster> del <host>
    proxymanager lb <cluster> restore <host>
    proxymanager lb <cluster> move <host1> <host2>
    proxymanager lb <cluster> add <host> <ip> 

proxymanager proxy
    proxymanager proxy new <hostname> 
        --ip <ip_address>
        --port <port>
        --k8s <cluster>
        --proxy-uri <uri>
        --ssl
        --ssl-bypass-firewall
        --proxy-ssl
        --proxy-verify-ssl-off
    proxymanager proxy list
        --k8s <cluster>
    proxymanager proxy enable <hostname>
        --k8s <cluster>
    proxymanager proxy disable <hostname>
        --k8s <cluster>
    proxymanager proxy remove <hostname>
        --k8s <cluster>

proxymanager ssl
    proxymanager ssl install <hostname>
        --bypass-firewall
    proxymanager ssl renew <hostname>
        --bypass-firewall

proxymanager fw
    proxymanager fw list
    proxymanager fw block <ip>
    proxymanager fw allow <ip>
