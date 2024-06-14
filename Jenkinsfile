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
            sh "git clone https://github.com/SomeoneWithOptions/api.diafestivo.co.git"
        }
    }
    stage('Build Binary'){
        steps{
            sh "cd api.diafestivo.co && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 /usr/local/go/bin/go build -o api"   
        }
    }
   } 
}