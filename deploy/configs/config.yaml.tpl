server:
  host: "0.0.0.0"
  port: 10020

database:
  host: "mysql"
  port: 3306
  user: "${MYSQL_USER}"
  password: "${MYSQL_PASSWORD}"
  dbname: "${MYSQL_DATABASE}"

jwt:
  secret: "${JWT_SECRET}"
  expire_hours: 24

acme:
  dns:
    resolvers: "8.8.8.8:53,1.1.1.1:53"
    timeout: "10s"

encryption:
  master_key: "${ENCRYPTION_MASTER_KEY}"
