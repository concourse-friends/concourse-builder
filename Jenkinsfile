pipeline {
   
    agent { 
        node { 
           label 'jenkins-slave' 
        } 
    } 
    
    environment {
        GOROOT = "/var/opt/go"
        GRDLHOME = "/var/opt/gradle"
        GOPATH = "$WORKSPACE"
        GOBIN = "${env.GOPATH}/bin"
        REPODIR = "src/github.com/peernova-private"
        PATH = "${env.GOROOT}/bin:${env.GRDLHOME}/bin:/usr/local/bin:/bin:/usr/bin:/usr/local/sbin:/usr/sbin:/sbin"
    }
    
    stages {
        
        stage('Prepare') { // Setup Check
            steps {
                sh '${GOROOT}/bin/go version'
                script {    
                    gver = sh (returnStdout: true, script: "git --version").trim()
                    echo "Git Version: ${gver}"
                }
                dir (env.GOBIN) {
                    deleteDir()
                }
            }
        }
        stage('Checkout') {
            steps {
                parallel (
                    "Concourse-Builder" : {
                        dir('src/github.com/peernova-private/concourse-builder') {
                            checkout scm
                        } 
                    }
		    "Integration" : {
                        dir("$REPODIR/integration") {
 			    checkout scm
                        }        
                    } 
                )
            }
    	}
	stage('pre-check') {
                parallel (
		     "Code-gen": {
				        
                            context="jenkins-ci/@code-gen"
                            setBuildStatus ("${context}", 'Jenkins CI build pending', 'PENDING')
                            dir("$GOPATH/$REPODIR/canary/ci/scripts") {
                                    sh './check_generation.sh'
                            }
            
                            if ('currentBuild.result = "SUCCESS"') {
                                githubNotify account: 'peernova-private', context: '${context}', credentialsId: 'peernova-go-token', description: 'Jenkins-ci build success', repo: 'canary', status: 'SUCCESS'   
                                //setBuildStatus ("${context}", 'Jenkins CI build success', 'SUCCESS')
            
                            } else {
                                echo 'Stage failed with an error'
                                githubNotify account: 'peernova-private', context: '${context}', credentialsId: 'peernova-go-token', description: 'Jenkins-ci build failed', repo: 'canary', status: 'FAILED'   

                                //setBuildStatus ("${context}", 'Jenkins CI build failed', 'FAILED')
                            }
                    },
                   "test": {
                        context="jenkins-ci/@test"
                        setBuildStatus ("${context}", 'Jenkins CI build pending', 'PENDING')
                         dir("$GOPATH/$REPODIR/cuneiform/src/tools/ci/scripts") {
                                    echo "Stage was successful"
                                
                        }
                            if('currentBuild.result = "SUCCESS"') {
                		githubNotify account: 'peernova-private', context: '${context}', credentialsId: 'peernova-go-token', description: 'Jenkins-ci build success', repo: 'canary', status: 'SUCCESS' 
                            } 
                            else {
			      githubNotify account: 'peernova-private', context: '${context}', credentialsId: 'peernova-go-token', description: 'Jenkins-ci build failed', repo: 'canary', status: 'FAILED'
                            }
                        
                    },    
                          
                    
                    "Lint-Test": {
			            
                                context="jenkins-ci/@lint"
                                setBuildStatus ("${context}", 'Jenkins CI build pending', 'PENDING')
                                dir("$GOPATH/$REPODIR/canary") {
                                    sh 'go get -u github.com/golang/lint/golint'
                                    sh 'golint ./...'
                                }
            
                                if ('currentBuild.result = "SUCCESS"') {
                                     githubNotify account: 'peernova-private', context: '${context}', credentialsId: 'peernova-go-token', description: 'Jenkins-ci build success', repo: 'canary', status: 'SUCCESS'
                               		 
                
                                } else {
                                    echo 'Stage failed with an error'
				    githubNotify account: 'peernova-private', context: '${context}', credentialsId: 'peernova-go-token', description: 'Jenkins-ci build failed', repo: 'canary', status: 'FAILED'
                                }
                        
			        },
                    "Copyright": {
				        
                            context="jenkins-ci/@copyright"
                            setBuildStatus ("${context}", 'Jenkins CI build pending', 'PENDING')
                            dir("$GOPATH/$REPODIR/canary/ci/scripts") {
                                    sh './check_copyright.sh'
                            }
            
                            if ('currentBuild.result = "SUCCESS"') {
                                 githubNotify account: 'peernova-private', context: '${context}', credentialsId: 'peernova-go-token', description: 'Jenkins-ci build success', repo: 'canary', status: 'SUCCESS'   
              			 
                            } else {
                                echo 'Stage failed with an error'
				githubNotify account: 'peernova-private', context: '${context}', credentialsId: 'peernova-go-token', description: 'Jenkins-ci build failed', repo: 'canary', status: 'FAILED'
                            }
                    },
                    "go_vendor": {
				        
                            context="jenkins-ci/@go-vendor"
                            setBuildStatus ("${context}", 'Jenkins CI build pending', 'PENDING')
                            dir("$GOPATH/$REPODIR/canary/ci/scripts") {
                                    sh './go_vendor.sh'
                            }
            
                            if ('currentBuild.result = "SUCCESS"') {
                                 githubNotify account: 'peernova-private', context: '${context}', credentialsId: 'peernova-go-token', description: 'Jenkins-ci build success', repo: 'canary', status: 'SUCCESS'   
           			 
                            } else {
                                echo 'Stage failed with an error'
				githubNotify account: 'peernova-private', context: '${context}', credentialsId: 'peernova-go-token', description: 'Jenkins-ci build failed', repo: 'canary', status: 'FAILED'
                            }
                    },
                    "Meta-Linter": {
				        
                            context="jenkins-ci/@meta-linter"
                            setBuildStatus ("${context}", 'Jenkins CI build pending', 'PENDING')
                            dir("$GOPATH/$REPODIR/canary/ci/scripts") {
                                    sh './meta_linter.sh'
                            }
            
                            if ('currentBuild.result = "SUCCESS"') {
                                 githubNotify account: 'peernova-private', context: '${context}', credentialsId: 'peernova-go-token', description: 'Jenkins-ci build success', repo: 'canary', status: 'SUCCESS'  
           			  
                            } else {
                                echo 'Stage failed with an error'
				githubNotify account: 'peernova-private', context: '${context}', credentialsId: 'peernova-go-token', description: 'Jenkins-ci build failed', repo: 'canary', status: 'FAILED'
                            }
                    },
                    "Vet" : {
                        context="jenkins-ci/@vet"
                        setBuildStatus ("${context}", 'Jenkins CI build pending', 'PENDING')
                        dir("$GOPATH/$REPODIR/canary") {
		                sh 'go vet  ./...'
                        }
                        if ('currentBuild.result = "SUCCESS"') {
                             githubNotify account: 'peernova-private', context: '${context}', credentialsId: 'peernova-go-token', description: 'Jenkins-ci build success', repo: 'canary', status: 'SUCCESS'           
           		 
                        } else {
                            echo 'Stage failed with an error'
			    githubNotify account: 'peernova-private', context: '${context}', credentialsId: 'peernova-go-token', description: 'Jenkins-ci build failed', repo: 'canary', status: 'FAILED'
                        }
                    }
                )
        }    
    } 
}   
