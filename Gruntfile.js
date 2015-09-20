module.exports = function(grunt) {
  grunt.loadNpmTasks('grunt-contrib-watch');
  grunt.loadNpmTasks('grunt-contrib-copy');
  grunt.loadNpmTasks('grunt-contrib-concat');
  grunt.loadNpmTasks('grunt-contrib-sass');
  grunt.loadNpmTasks('grunt-contrib-cssmin');
  grunt.loadNpmTasks('grunt-contrib-uglify');

  grunt.initConfig({
    watch: {
      sass: {
        files: ['assets/src/sass/**/*.scss'],
        tasks: ['sass', 'concat', 'cssmin']
      },
      js: {
        files: ['assets/src/js/**/*.js'],
        tasks: ['uglify:main']
      },
    },
    sass: {
      dist: {
        options: {
          style: 'expanded',
          sourcemap: 'none'
        },
        files: [{
          expand: true,
          cwd: 'assets/src/sass',
          src: ['**/*.scss'],
          dest: 'temp/css',
          ext: '.css'
        }]
      }
    },
    concat: {
      css: {
        src: ['node_modules/normalize.css/normalize.css',
          'node_modules/font-awesome/css/font-awesome.css',
          'node_modules/animate.css/animate.min.css',
          'node_modules/codemirror/lib/codemirror.css',
          'node_modules/codemirror/theme/mdn-like.css',
          'assets/src/js/highlight/styles/monokai_sublime.css',
          'temp/css/**/*.css'
        ],
        dest: 'temp/css/main.css',
      },
    },
    copy: {
      main: {
        files: [{
          expand: true,
          flatten: true,
          src: ['node_modules/font-awesome/fonts/**'],
          dest: 'assets/fonts'
        }],
      },
    },
    cssmin: {
      target: {
        files: [{
          expand: true,
          cwd: 'temp/css/',
          src: ['*.css', '!*.min.css'],
          dest: 'assets/css/',
          ext: '.min.css'
        }]
      }
    },
    uglify: {
      plugins: {
        files: {
          'assets/js/plugins.min.js': ['node_modules/jquery/dist/jquery.min.js',
            'node_modules/perfect-scrollbar/dist/js/min/perfect-scrollbar.jquery.min.js',
            'node_modules/showdown/dist/showdown.min.js',
            'node_modules/noty/js/noty/packaged/jquery.noty.packaged.min.js',
            'node_modules/jquery-pjax/jquery.pjax.js',
            'node_modules/jquery-serializejson/jquery.serializejson.min.js',
            'node_modules/codemirror/lib/codemirror.js',
            'node_modules/codemirror/mode/css/css.js',
            'node_modules/codemirror/mode/javascript/javascript.js',
            'node_modules/codemirror/mode/markdown/markdown.js',
            'node_modules/codemirror/mode/sass/sass.js',
            'node_modules/codemirror/mode/htmlmixed/htmlmixed.js',
            'assets/src/js/highlight/highlight.pack.js',
            'node_modules/js-cookie/src/js.cookie.js'
          ]
        }
      },
      main: {
        files: {
          'assets/js/app.min.js': ['assets/src/js/**/*.js']
        }
      }
    }
  });

  grunt.registerTask('default', ['copy', 'sass', 'concat', 'cssmin', 'uglify']);
};