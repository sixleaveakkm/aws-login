# aws-login

aws-login is a golang CLI script helps configure aws cli users who are required using MFA.

## Todo
### Platform Support
- [x] unix (include MacOS)
- [ ] Windows

### Functions
- [x] mfa
    - [x] profile generate
    - [x] bash completion
- [x] role
    - [x] profile generate
    - [x] bash completion
- [x] login

### Other
- [ ] Support aws configure profile  
`aws configure` generates static profile in `.config`

- [ ] prompt CUI while some key parameter is missing
  - [ ] add a flag to not prompt
 
- [ ] Hardware MFA device test?
 

## Background
For security reason, some developers are required using MFA to access to AWS account.

This script is aiming for these people to configure and login cli with mfa, and assume role with mfa.

##  How To Install
If you have `Golang` installed, you can simply do
```bash
go install github.com/sixleaveakkm/aws-login.git
```


### zsh completion (Optional)
1. with [zsh-completions](https://github.com/zsh-users/zsh-completions)  
If you are using `zsh-completions`, 

### bash completion (Optional)
> todo

## How to use it
1. Set mfa config  
use `aws-login config mfa` or `aws-login config role` to config profile  

2. Login
`aws-login --profile <your-profile-name> YOUCOD`
 
3. Done

4. Extra (`aws-export`)
### zsh
Add following code into your `.zshrc` file  
After execute `aws-login -p test 123456`,  
use `aws-export test` to export the profile.
  
```zsh
function zsh-export(){
    export AWS_PROFILE="$1" && export AWS_REGION=ap-northeast-1
}
```
