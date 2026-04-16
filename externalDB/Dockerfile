FROM mysql:8.0

# Sätt miljövariabler för databasen
ENV MYSQL_ROOT_PASSWORD=password123
ENV MYSQL_DATABASE=industrial_db
ENV MYSQL_USER=sensor_user
ENV MYSQL_PASSWORD=password123

# Lägg till ett SQL-skript i mappen /docker-entrypoint-initdb.d/
# Alla .sql-filer i denna mapp körs automatiskt när containern startas första gången.
RUN echo "CREATE TABLE sensor_data (id INT AUTO_INCREMENT PRIMARY KEY, timestamp DATETIME, machine_id VARCHAR(50), temperature INT, pressure INT);" > /docker-entrypoint-initdb.d/init.sql

EXPOSE 3306
