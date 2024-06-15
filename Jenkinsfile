pipeline{
   agent any

    stages {
    stage('Clean Up'){
        steps {
           deleteDir()
           sh '''
            docker system prune -f 
            docker rmi $(docker images -q)
            docker system prune -f 
           '''
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
                sh '''
                    export PATH=$PATH:/usr/local/go/bin
                    go mod download
                    go test -v ./...
                '''
            }
        }
    }

    stage('Build Binaries'){
        steps{
            sh "cd api.diafestivo.co && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 /usr/local/go/bin/go build -o apiAMD"   
            sh "cd api.diafestivo.co && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 /usr/local/go/bin/go build -o apiARM"
        }
    }

    stage('Build Docker Image'){
        steps{
            sh "cd api.diafestivo.co && docker build -t api.diafestivo.co:${env.BUILD_NUMBER} ."
        }
    }

   } 
}