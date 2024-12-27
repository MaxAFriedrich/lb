# Use the latest Alpine image
FROM alpine:latest

# Install OpenSSH
RUN apk add --no-cache openssh

# Create a user with username 'max' and password 'max'
RUN adduser -D max && echo "max:max" | chpasswd

# Set up SSH
RUN ssh-keygen -A

# Expose the SSH port
EXPOSE 22

# Start the OpenSSH server
CMD ["/usr/sbin/sshd", "-D"]
