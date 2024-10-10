#!/bin/bash

set -e

echo -e "\n**** Domains to VPN (via dns server and ipset) \xF0\x9F\x98\x83 ***\n"

while true; do
    read -p "* Скрипту требуются пакеты: wireguard/wireguard-tools, curl, ipset, iptables и iproute2 - они установлены? (y/n) "  yn
    case $yn in
        [YyНн]* ) break;;
        [NnТт]* ) echo -e "\n*** Установишь, возвращайся *** \n"; exit;;
        * ) echo "Ну y или n в чем проблема?";;
    esac
done


echo "* Скачиваю конфиг сервиса"
sudo curl -Lks -z /etc/systemd/system/dns-ipset.service https://github.com/xMlex/dns-ipset/raw/refs/heads/main/dns-ipset.service -o /etc/systemd/system/dns-ipset.service
sudo systemctl daemon-reload

if [ ! -d /etc/wireguard ]; then echo "Директория /etc/wireguard не существует - паника 🫠"; exit 1; fi

sudo mkdir -p /opt/dns-ipset
sudo curl -Lks https://github.com/xMlex/dns-ipset/raw/refs/heads/main/wg-dns-ipset.example.conf -o /etc/wireguard/wg-dns-ipset.example.conf
sudo curl -Lks https://github.com/xMlex/dns-ipset/raw/refs/heads/main/config.example.yaml -o /opt/dns-ipset/config.example.yaml

if [ ! -f /opt/dns-ipset/config.yaml ]; then echo "Базовый конфиг установлен - /opt/dns-ipset/config.yaml"; sudo cp /opt/dns-ipset/config.example.yaml /opt/dns-ipset/config.yaml; fi

echo "* Скачиваю dns-ipset"
sudo curl -L -z /opt/dns-ipset/dns-ipset-linux-amd64.tar.gz -o /opt/dns-ipset/dns-ipset-linux-amd64.tar.gz https://github.com/xMlex/dns-ipset/releases/download/v1.0.0/dns-ipset-linux-amd64.tar.gz
sudo tar -C /opt/dns-ipset -xzf /opt/dns-ipset/dns-ipset-linux-amd64.tar.gz
sudo mv /opt/dns-ipset/dns-ipset-linux-amd64 /opt/dns-ipset/dns-ipset

echo "* Включаю службу dns-ipset"
sudo systemctl enable dns-ipset
echo "* Стартую службу dns-ipset"
sudo systemctl restart dns-ipset

echo ""
echo -e "**** Поздравляю \xF0\x9F\x98\x83 ***"
echo "* Список доменов которые должны быть доступны через VPN редактируй тут - /opt/dns-ipset/config.yaml (секция ipsets->vpn )"
echo "* На основе этого конфига скорректируй совй wg0.conf - /etc/wireguard/wg-dns-ipset.example.conf"
echo "* Перезапусти WG, systemctl restart dns-ipset и радуйся доступу к запрещёнке"
echo ""