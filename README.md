# Köra experimentet

Allt körs via ett script.

## Körning

Stå i root-mappen för repot och kör:

```bash
chmod +x start_test.sh
./start_test.sh

Scriptet:

installerar dependencies vid behov
startar docker och minikube
bygger alla images
deployar applikationerna i Kubernetes
kör experimentet
sparar logg i experiment_log.txt
kör analys
städar upp resurser
Första körningen

Om docker installeras första gången kommer scriptet att avbrytas efter att ha lagt till din användare i docker-gruppen.

Logga då ut och in igen, och kör scriptet en gång till:

./start_test.sh
Output

Efter körning:

experiment_log.txt innehåller loggar från testet
analys skrivs ut i terminalen
Om något inte fungerar

Kontrollera:

minikube status
kubectl get pods
docker ps

Vanliga problem:

minikube startar inte → kör minikube delete och testa igen
pods fastnar → kontrollera med kubectl describe pod
port 3306 är upptagen → stoppa lokal MySQL
