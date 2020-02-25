# aws-login

aws-login is a golang CLI script helps configure aws cli users who are required using MFA.

## Features
### Platform Support
- [x] unix (include MacOS)
- [ ] Windows

### Functions
- [ ] mfa
    - [ ] profile generate
    - [ ] bash completion
- [ ] role
    - [ ] profile generate
    - [ ] bash completion
- [ ] login


## Background
For security reason, some developers are required using MFA to access to AWS account.

This script is aiming for these people to configure and login cli with mfa, and assume role with mfa.

##  How To Install
```bash
go install github.com/sixleaveakkm/aws-login.git
```

### zsh completion (Optional)
1. with [zsh-completions](https://github.com/zsh-users/zsh-completions)  
If you are using `zsh-completions`, 

### bash completion (Optional)

## How to use it
