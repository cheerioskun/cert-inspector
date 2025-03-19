#!/bin/bash
set -e

# Create testdata directory structure
mkdir -p testdata/certs/{valid,expired,duplicates}
mkdir -p testdata/certs/valid/{webapp,database,api}
mkdir -p testdata/certs/nested/deeper/evendeeper

# Function to create a self-signed certificate
create_cert() {
    local name=$1
    local days=$2
    local dir=$3
    
    # Generate private key
    openssl genrsa -out "${dir}/${name}.key" 2048
    
    # Create config file
    cat > "${dir}/${name}.cnf" << EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
CN = ${name}.example.com

[v3_req]
keyUsage = keyEncipherment, dataEncipherment, digitalSignature
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = ${name}.example.com
DNS.2 = ${name}-backup.example.com
EOF

    # Generate certificate
    openssl req -new -x509 -key "${dir}/${name}.key" -out "${dir}/${name}.crt" \
        -days $days -config "${dir}/${name}.cnf"
    
    # Also create a PEM format with both key and cert for some certs
    if [ $(($RANDOM % 2)) -eq 0 ]; then
        cat "${dir}/${name}.key" "${dir}/${name}.crt" > "${dir}/${name}.pem"
    fi
    
    # Generate a certificate signing request for some certs
    if [ $(($RANDOM % 2)) -eq 0 ]; then
        openssl req -new -key "${dir}/${name}.key" -out "${dir}/${name}.csr" \
            -config "${dir}/${name}.cnf"
    fi
}

# Create valid certificates
create_cert "webapp-frontend" 365 "testdata/certs/valid/webapp"
create_cert "webapp-backend" 365 "testdata/certs/valid/webapp"
create_cert "database-primary" 365 "testdata/certs/valid/database"
create_cert "database-replica" 365 "testdata/certs/valid/database"
create_cert "api-gateway" 365 "testdata/certs/valid/api"
create_cert "api-internal" 365 "testdata/certs/valid/api"

create_cert "almost-expired" 10 "testdata/certs/expired"

# Create duplicate certificates
create_cert "duplicate-cert" 365 "testdata/certs/duplicates"
# Copy the same certificate to different locations to simulate duplicates
cp testdata/certs/duplicates/duplicate-cert.crt testdata/certs/nested/duplicate-cert.crt
cp testdata/certs/duplicates/duplicate-cert.key testdata/certs/nested/duplicate-cert.key

# Create certificates in nested directories
create_cert "nested-service" 365 "testdata/certs/nested"
create_cert "deeper-service" 365 "testdata/certs/nested/deeper"
create_cert "deepest-service" 365 "testdata/certs/nested/deeper/evendeeper"

# Create a certificate chain
cat > create_chain.sh << 'EOF'
#!/bin/bash

# Create a root CA
mkdir -p testdata/certs/chain
cd testdata/certs/chain

# Generate root CA private key
openssl genrsa -out root-ca.key.pem 4096

# Generate root CA certificate
openssl req -x509 -new -nodes -key root-ca.key.pem -sha256 -days 1024 -out root-ca.crt.pem \
    -subj "/C=US/ST=State/L=City/O=Organization/OU=Unit/CN=Example Root CA"

# Create intermediate CA
mkdir -p intermediate
openssl genrsa -out intermediate/intermediate-ca.key.pem 4096

# Create intermediate CA CSR
openssl req -new -key intermediate/intermediate-ca.key.pem -out intermediate/intermediate-ca.csr.pem \
    -subj "/C=US/ST=State/L=City/O=Organization/OU=Unit/CN=Example Intermediate CA"

# Create intermediate CA config
cat > intermediate/intermediate-ca.cnf << EOT
[v3_intermediate_ca]
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:true, pathlen:0
keyUsage = critical, digitalSignature, cRLSign, keyCertSign
EOT

# Sign intermediate CA with root CA
openssl x509 -req -in intermediate/intermediate-ca.csr.pem \
    -CA root-ca.crt.pem -CAkey root-ca.key.pem -CAcreateserial \
    -out intermediate/intermediate-ca.crt.pem -days 730 \
    -sha256 -extfile intermediate/intermediate-ca.cnf -extensions v3_intermediate_ca

# Create leaf cert
mkdir -p leaf
openssl genrsa -out leaf/example-service.key.pem 2048

# Create leaf cert CSR
openssl req -new -key leaf/example-service.key.pem -out leaf/example-service.csr.pem \
    -subj "/C=US/ST=State/L=City/O=Organization/OU=Unit/CN=example-service.example.com"

# Create leaf cert config
cat > leaf/example-service.cnf << EOT
[v3_leaf]
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:false
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth, clientAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = example-service.example.com
DNS.2 = example-service-backup.example.com
EOT

# Sign leaf cert with intermediate CA
openssl x509 -req -in leaf/example-service.csr.pem \
    -CA intermediate/intermediate-ca.crt.pem -CAkey intermediate/intermediate-ca.key.pem -CAcreateserial \
    -out leaf/example-service.crt.pem -days 365 \
    -sha256 -extfile leaf/example-service.cnf -extensions v3_leaf

# Create complete chain file
cat leaf/example-service.crt.pem intermediate/intermediate-ca.crt.pem root-ca.crt.pem > leaf/example-service.chain.pem
EOF

chmod +x create_chain.sh
./create_chain.sh

# Create a test script to run cert-inspector
cat > test_cert_inspector.sh << 'EOF'
#!/bin/bash

# Build cert-inspector if needed
# Uncomment the following line if you want to build from source
# go build -o cert-inspector ./...

echo "Running cert-inspector search on testdata directory..."
./cert-inspector search ./testdata

echo -e "\nRunning cert-inspector tree to visualize the certificates..."
./cert-inspector tree

echo -e "\nTest completed. Check the output above for the certificate information."
EOF

chmod +x test_cert_inspector.sh

echo "Test data directory created successfully!"
echo "Run ./test_cert_inspector.sh to test cert-inspector on the generated data."