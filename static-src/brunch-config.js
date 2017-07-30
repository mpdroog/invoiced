module.exports = {
  files: {
    javascripts: {
      joinTo: {
        'vendor.js': /^(?!app)/,
        'app.js': /^app/
      }
    },
    stylesheets: {
      joinTo: 'app.css'
    }
  },

  paths: {
    public: "../static/assets"
  },

  modules: {
    autoRequire: {
      'app.js': ['main']
    }
  },

  plugins: {
    postcss: {
      processors: [
        require('csswring')()
      ]
    }
  }
};
