pipeline{
   agent any

   stages {
    stage('cleaunUp'){
        steps {
           deleteDir()
        }
    }
    stage('Clone Repo'){
        steps{
            sh "git clone --branch deploy https://github.com/SomeoneWithOptions/api.diafestivo.co.git"
        }
    }

    stage ("Test"){
        environment {
            REDIS_DB = credentials('REDIS_DB')
            IP_INFO_TOKEN = credentials('IP_INFO_TOKEN')
        }
        steps{
            dir("api.diafestivo.co"){
                sh "/usr/local/go/bin/go mod download"
                sh "/usr/local/go/bin/go mod tidy"
                sh "/usr/local/go/bin/go test -v ./..."
                sh "pwd"
            }
        }
    }

    stage('Build Binary'){
        steps{
            sh "cd api.diafestivo.co && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 /usr/local/go/bin/go build -o api"   
        }
    }

    stage ('Build Docker Image'){
        steps{
            sh "docker build . -t api.diafestivo.co:latest"
        }
    }
   } 
}