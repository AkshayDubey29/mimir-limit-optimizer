// Placeholder JavaScript file for go:embed directive
// This ensures the ui/build/static/js directory has embeddable files
// Real JS files will overwrite this during the build process

console.log('Mimir Limit Optimizer - Placeholder UI');
console.log('The real UI will be built during the Docker build process');

// Simple placeholder functionality
document.addEventListener('DOMContentLoaded', function() {
    const root = document.getElementById('root');
    if (root) {
        root.innerHTML = `
            <div class="placeholder-message">
                <h1>Mimir Limit Optimizer</h1>
                <p>Placeholder UI - Real dashboard will be available after build</p>
                <p>This is a development/CI placeholder</p>
            </div>
        `;
    }
});

// Export placeholder module for compatibility
if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        placeholder: true,
        message: 'UI build placeholder'
    };
} 