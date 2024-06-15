pipeline{
   agent any

    parameters{
            booleanParam(name: 'RUN_TESTS', defaultValue: true, description: 'Run tests?')
            choice(name: 'AWS_REGION', choices: ['us-east-1', 'us-east-2', 'us-west-1', 'us-west-2'], description: 'Select AWS Region')
        }

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
                sh '''
                    export PATH=$PATH:/usr/local/go/bin
                    go mod download
                    go test -v ./...
                    pwd
                '''
            }
        }
    }

    stage('Build Binary'){
        steps{
            sh "cd api.diafestivo.co && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 /usr/local/go/bin/go build -o api"   
        }
    }

   } 
}