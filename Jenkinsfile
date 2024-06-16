pipeline{
   agent any

   environment{
    AWS_CONTAINER = '745912973548.dkr.ecr.us-east-1.amazonaws.com'
   }

    stages {
    stage('Clean Up'){
        steps {
           deleteDir()
           sh '''
            docker system prune -f 
            docker rmi -f $(docker images -q)
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

    stage('Build Docker Image'){
        steps{
            dir("api.diafestivo.co"){
                sh "aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin ${AWS_CONTAINER}"
                sh "docker build . -t ${AWS_CONTAINER}/api-diafestivo:latest"
                sh "docker push ${AWS_CONTAINER}/api-diafestivo:latest"
            }
        }
    }

   } 
}