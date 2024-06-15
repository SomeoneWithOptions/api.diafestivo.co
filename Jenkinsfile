pipeline{
   agent any

    stages {
    stage('Clean Up'){
        steps {
           deleteDir()
        }
    }
    stage('Clone Repo'){
        steps{
            sh "git clone --branch deploy https://github.com/SomeoneWithOptions/api.diafestivo.co.git"
        }
    }

    stage ("Test Code"){
       
        environment {
            REDIS_DB = credentials('REDIS_DB')
            IP_INFO_TOKEN = credentials('IP_INFO_TOKEN')
        }
        steps{
            dir("api.diafestivo.co"){
                sh "echo ${env.BUILD_NUMBER}"
                sh '''
                    export PATH=$PATH:/usr/local/go/bin
                    go mod download
                    go test -v ./...
                '''
            }
        }
    }

    stage('Build Binary'){
        steps{
            sh "cd api.diafestivo.co && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 /usr/local/go/bin/go build -o api"   
        }
    }

    stage('Build Docker Image'){
        steps{
            sh "cd api.diafestivo.co && docker build -t api.diafestivo.co:${env.BUILD_NUMBER} ."
        }
    }

   } 
}