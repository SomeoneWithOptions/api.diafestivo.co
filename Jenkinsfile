pipeline{
   agent any

   environment{
    AWS_CONTAINER = '745912973548.dkr.ecr.us-east-1.amazonaws.com'
   }

    stages {
    stage('Clean Up'){
        steps {
           deleteDir()
           script {
                    def imageCount = sh(script: 'docker images -q 2> /dev/null | wc -l', returnStdout: true).trim()
                    if (imageCount.toInteger() > 0) {
                        sh 'docker system prune -f'
                        sh 'docker rmi -f $(docker images -q)'
                        sh 'docker system prune -f'
                    } else {
                        echo 'No images to remove.'
                    }
        }
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

    stage('Build and Upload to AWS'){
        steps{
            dir("api.diafestivo.co"){
                sh "aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin ${AWS_CONTAINER}"
                sh "docker build . -t ${AWS_CONTAINER}/api-diafestivo:${env.BRANCH_NAME}"
                sh "docker push ${AWS_CONTAINER}/api-diafestivo:${env.BRANCH_NAME}"
            }
        }
    }

   } 
}