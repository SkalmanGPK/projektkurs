# Köra experimentet lokalt

Det här projektet är uppsatt så att hela miljön kan startas och köras via ett enda script.

## Förutsättningar

Följande måste vara installerat och fungera:

- docker
- kubectl
- minikube
- python3

Kontrollera t.ex.:

```bash
docker --version
kubectl version --client
minikube version
python3 --version

Minikube behöver kunna starta (t.ex. med docker driver).

Körning

Stå i root-mappen för repot och kör:

chmod +x run.sh
./run.sh

Scriptet gör följande:

Startar minikube (om det inte redan kör)
Bygger alla images
Startar extern databas (container)
Deployar alla applikationer i Kubernetes
Väntar tills pods är redo
Kör experimentet i tre faser:
normal drift
databasen stängs av
databasen startas igen
Samlar loggar i experiment_log.txt
Kör analys via analyze.py
Städar upp alla resurser
Output

Efter körning finns:

experiment_log.txt – rå logg från testet
output från analys-scriptet i terminalen

Vanliga problem

Om något failar:

Kör minikube status och säkerställ att den är igång
Kör kubectl get pods och se om något fastnar i Pending/CrashLoop
Kontrollera att port 3306 inte redan används (databasen)
Köra om testet

Det går att köra scriptet igen direkt:

./run.sh

Scriptet tar bort tidigare resurser själv.
