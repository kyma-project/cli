#!/usr/bin/env groovy
def label = "kyma-${UUID.randomUUID().toString()}"
def application = 'kymactl'

dockerRegistry = 'eu.gcr.io/kyma-project/'
def dockerCredentials = 'gcr-rw'

def isMaster = env.BRANCH_NAME == 'master'
def appVersion = env.TAG_NAME?env.TAG_NAME:"develop-${env.BRANCH_NAME}"


echo """
********************************
Job started with the following parameters:
BRANCH_NAME=${env.BRANCH_NAME}
TAG_NAME=${env.TAG_NAME}
********************************
"""

podTemplate(label: label) {
    node(label) {
        try {
            timestamps {
                timeout(time:20, unit:"MINUTES") {
                    ansiColor('xterm') {
                        stage("setup") {
                            checkout scm

                            withCredentials([usernamePassword(credentialsId: dockerCredentials, passwordVariable: 'pwd', usernameVariable: 'uname')]) {
                                sh "docker login -u $uname -p '$pwd' $dockerRegistry"
                            }
                        }

                        stage("install dependencies $application") {
                            execute("make resolve")
                        }

                        stage("code quality $application") {
                            execute("make validate")
                        }

                        stage("build $application") {
                            execute("go generate ./...")
                            def flags = "-ldflags \"-X github.com/kyma-incubator/kymactl/cmd.Version=${appVersion}\""
                            execute("CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./bin/kymactl.exe ${flags}")
                            execute("CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/kymactl-linux ${flags}")
                            execute("CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ./bin/kymactl-darwin ${flags}")
                        }

                        stage("test $application") {
                            //execute("make test-report")
                            //junit '**/unit-tests.xml'
                        }

                        stage("integration test $application") {
                            sh "bin/kymactl-linux help"
                        }


                        stage("archive $application") {
                            archiveArtifacts artifacts: 'bin/*', onlyIfSuccessful: true
                        }
                    }
                }
            }
        } catch (ex) {
            echo "Got exception: ${ex}"
            currentBuild.result = "FAILURE"
            def body = "${currentBuild.currentResult} ${env.JOB_NAME}${env.BUILD_DISPLAY_NAME}: on branch: ${params.GIT_BRANCH}. See details: ${env.BUILD_URL}"
            emailext body: body, recipientProviders: [[$class: 'DevelopersRecipientProvider'], [$class: 'CulpritsRecipientProvider'], [$class: 'RequesterRecipientProvider']], subject: "${currentBuild.currentResult}: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]'"
        }
    }
}

def execute(command, envs = []) {
    def buildpack = 'golang-buildpack:0.0.9'
    def repositoryName = 'kymactl'
    def envText = envs=='' ? '' : "--env $envs"
    workDir = pwd()
    sh "docker run --rm -v $workDir:/go/src/github.com/kyma-incubator/$repositoryName/ -w /go/src/github.com/kyma-incubator/$repositoryName $envText ${dockerRegistry}$buildpack /bin/bash -c '$command'"
}