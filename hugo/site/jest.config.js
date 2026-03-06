module.exports = {
  testEnvironment: 'jsdom',
  testMatch: ['**/static/js/**/__tests__/**/*.js', '**/static/js/**/?(*.)+(spec|test).js'],
  collectCoverageFrom: [
    'static/js/**/*.js',
    '!static/js/**/*.test.js'
  ],
  coverageDirectory: 'coverage',
  verbose: true
};
