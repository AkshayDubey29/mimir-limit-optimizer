# Contributing to Mimir Limit Optimizer

Thank you for your interest in contributing to Mimir Limit Optimizer! This document provides guidelines and information for contributors.

## 🚀 Getting Started

### Prerequisites

- **Go**: 1.21 or later
- **Docker**: For building container images
- **Kubernetes**: 1.24+ for testing
- **Helm**: 3.0+ for deployment testing
- **Git**: For version control

### Development Setup

1. **Fork and Clone**
   ```bash
   git clone https://github.com/your-username/mimir-limit-optimizer.git
   cd mimir-limit-optimizer
   ```

2. **Install Dependencies**
   ```bash
   go mod download
   go mod tidy
   ```

3. **Build the Project**
   ```bash
   make build
   ```

4. **Run Tests**
   ```bash
   make test
   ```

5. **Run Locally**
   ```bash
   go run main.go --config=config.yaml --log-level=debug
   ```

## 🎯 Ways to Contribute

### 🐛 Bug Reports

When reporting bugs, please include:

- **Clear Description**: What did you expect vs. what happened?
- **Environment**: Kubernetes version, Mimir version, deployment details
- **Reproduction Steps**: Minimal steps to reproduce the issue
- **Logs**: Relevant log excerpts (use `kubectl logs`)
- **Configuration**: Sanitized configuration files
- **Version**: Mimir Limit Optimizer version

### ✨ Feature Requests

For new features, please provide:

- **Use Case**: Why is this feature needed?
- **Proposed Solution**: How should it work?
- **Alternatives**: Other solutions you've considered
- **Impact**: Who would benefit from this feature?

### 📝 Documentation

Help improve our documentation:

- **README**: Updates and clarifications
- **Code Comments**: Inline documentation
- **Examples**: Usage examples and tutorials
- **Architecture**: Design documents and diagrams

### 💻 Code Contributions

We welcome code contributions! Please follow our development process.

## 🔄 Development Process

### 1. Create an Issue
   - Discuss your changes before implementing
   - Get feedback from maintainers
   - Ensure alignment with project goals

### 2. Fork and Branch
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/issue-number-description
   ```

### 3. Implement Changes
   - Follow coding standards
   - Add tests for new functionality
   - Update documentation as needed
   - Ensure all tests pass

### 4. Test Thoroughly
   ```bash
   # Run all tests
   make test
   
   # Run linting
   make lint
   
   # Test in real environment
   make test-integration
   ```

### 5. Commit and Push
   ```bash
   git add .
   git commit -m "feat: add new cost optimization algorithm"
   git push origin feature/your-feature-name
   ```

### 6. Create Pull Request
   - Use clear descriptions
   - Link related issues
   - Include testing evidence

## 📋 Coding Standards

### Go Code Style

- **gofmt**: Code must be formatted with `gofmt`
- **golint**: Follow Go linting recommendations
- **govet**: Code must pass `go vet`
- **Naming**: Use clear, descriptive names
- **Comments**: Public functions and types must have comments
- **Error Handling**: Always handle errors explicitly

### Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Test additions/changes
- `chore`: Maintenance tasks

**Examples:**
```
feat(costcontrol): add budget enforcement with alerts
fix(circuitbreaker): resolve race condition in state management
docs(readme): update installation instructions
```

## 🔒 Security

### Reporting Security Issues

**DO NOT** create public issues for security vulnerabilities.

Instead:
1. Email akshaydubey2912@gmail.com
2. Include "SECURITY" in the subject line
3. Provide detailed description
4. Include reproduction steps if possible

## 📞 Getting Help

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and discussions
- **Email**: akshaydubey2912@gmail.com for maintainer contact

## 🙏 Thank You

Thank you for contributing to Mimir Limit Optimizer! Your contributions help make observability better for everyone. 