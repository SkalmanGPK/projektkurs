Dokumentation – IDS Sensor (App2)
Översikt

Denna applikation fungerar som en enkel Intrusion Detection Sensor (IDS) som körs i ett Kubernetes-kluster. Syftet med applikationen är att övervaka inkommande HTTP-trafik och identifiera potentiellt skadliga beteenden, exempelvis:

misstänkta User-Agent headers
ovanliga query strings
onormalt hög trafikvolym som kan indikera en DoS-attack

Applikationen är skriven i Go och körs i en Docker-container som distribueras i Kubernetes.

Den exponeras internt i klustret via en ClusterIP-service, vilket innebär att den endast kan nås av andra pods i samma kluster.

Systemarkitektur

IDS-sensorn fungerar som en del av ett större system där flera applikationer kommunicerar inom ett Kubernetes-kluster.

Grundläggande flöde:

App1 / annan tjänst
        │
        │ HTTP request
        ▼
IDS Sensor (App2)
        │
        │ analyserar trafik
        ▼
Loggar potentiella attacker

IDS-sensorn analyserar trafiken i realtid och loggar misstänkta aktiviteter till standard output, vilket i Kubernetes kan läsas via:

kubectl logs
Funktion

Applikationen består av tre huvudsakliga funktioner:

1. HTTP trafikövervakning

IDS:en startar en HTTP-server på port 8080 och lyssnar på endpointen:

/monitor

När en request skickas till denna endpoint utförs trafikinspektion.

2. Header-inspektion

Applikationen kontrollerar HTTP-headern:

User-Agent

Om ett känt attackverktyg identifieras loggas detta.

Exempel:

User-Agent: AttackTool/1.0

Logg:

Malicious User-Agent detected: AttackTool/1.0 from IP 10.1.0.15

Detta simulerar hur en IDS kan identifiera trafik från kända attackverktyg.

3. Simulerad Deep Packet Inspection (DPI)

IDS:en analyserar även URL-query strings.

Exempel:

/monitor?cmd=delete

Query-strängen loggas eftersom den kan indikera attacker som:

command injection
SQL injection
scanning

Loggexempel:

Located suspicious query string: cmd=delete
4. DoS-detektion

Applikationen övervakar antalet inkommande requests under ett 10 sekunders intervall.

Om mer än 50 requests tas emot under denna period loggas en varning.

Loggexempel:

Traffic spike detected! 75 requests in 10s. Possible DoS attack.

Detta simulerar en enkel mekanism för att upptäcka Denial-of-Service attacker.

Implementation
1. Go-applikationen

Applikationen använder följande Go-bibliotek:

Bibliotek	Funktion
net/http	HTTP-server
fmt	loggutskrift
time	tidsbaserad DoS-detektion

Servern startas med:

http.ListenAndServe(":8080", nil)
Containerisering

Applikationen körs i en Docker-container.

Dockerfile

Dockerfilen bygger applikationen i en lättviktig Alpine-baserad Go-miljö.

Steg:

Basimage
FROM golang:1.22-alpine

En minimal Go-miljö används för att hålla containerstorleken liten.

Arbetskatalog
WORKDIR /app

Alla filer kopieras till katalogen /app.

Kopiera källkod
COPY main.go .

Go-filen kopieras in i containern.

Inaktivera Go modules
ENV GO111MODULE=off

Eftersom projektet endast består av en fil behövs inga externa moduler.

Bygg applikationen
RUN go build -o ids main.go

Go-koden kompileras till en binär fil som heter:

ids
Exponera port
EXPOSE 8080

Containern deklarerar att den använder port 8080.

Startkommando
CMD ["./ids"]

När containern startas körs IDS-programmet.

Kubernetes Deployment

Applikationen exponeras i Kubernetes med en Service.

app2-deployment.yaml

Denna fil skapar en Kubernetes-service.

API Version
apiVersion: v1

Använder Kubernetes core API.

Resurstyp
kind: Service

Definierar en nätverksservice.

Metadata
metadata:
  name: ids-app2-service

Servicen får namnet:

ids-app2-service
Pod-selektor
selector:
  app: ids-app2

Servicen kopplas till pods som har label:

app=ids-app2
Portmapping
ports:
  - protocol: TCP
    port: 80
    targetPort: 8080

Det innebär:

Service Port	Container Port
80	8080

All trafik till port 80 skickas vidare till applikationen på port 8080.

Servicetyp
type: ClusterIP

Detta innebär att tjänsten endast är tillgänglig inom Kubernetes-klustret.

Den kan inte nås direkt från internet.

Implementation steg
1. Bygg Docker-imagen
docker build -t ids-app2 .
2. Exportera image (valfritt)
docker save ids-app2 > ids-app2.tar
3. Importera image i MicroK8s
microk8s ctr image import ids-app2.tar
4. Skapa deployment
kubectl apply -f app2-deployment.yaml
5. Kontrollera pods
kubectl get pods
6. Kontrollera service
kubectl get services
7. Visa loggar
kubectl logs <pod-namn>

Här visas IDS-detektionerna.

Exempel på test
Normal trafik
curl http://ids-app2-service/monitor
Misstänkt User-Agent
curl -A "AttackTool/1.0" http://ids-app2-service/monitor
Misstänkt query
curl http://ids-app2-service/monitor?cmd=delete
Sammanfattning

Denna IDS-sensor demonstrerar hur ett säkerhetsverktyg kan implementeras i en containerbaserad Kubernetes-miljö.

Den visar tre centrala säkerhetsfunktioner:

Header-analys
Deep Packet Inspection (simulerad)
DoS-detektion via trafikmätning

Applikationen är avsiktligt enkel men illustrerar hur IDS-funktionalitet kan integreras i ett mikrotjänstsystem.
