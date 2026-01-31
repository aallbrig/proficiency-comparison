module.exports = {
  testEnvironment: 'jsdom',
  testMatch: ['**/__tests__/**/*.js', '**/?(*.)+(spec|test).js'],
  collectCoverageFrom: [
    'static/js/**/*.js',
    '!static/js/**/*.test.js'
  ],
  coverageDirectory: 'coverage',
  verbose: true
};
