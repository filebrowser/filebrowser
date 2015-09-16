module.exports = function(grunt) {

  grunt.loadNpmTasks('grunt-contrib-watch');
  grunt.loadNpmTasks('grunt-contrib-copy');
  grunt.loadNpmTasks('grunt-contrib-concat');
  grunt.loadNpmTasks('grunt-contrib-sass');
  grunt.loadNpmTasks('grunt-contrib-cssmin');
  grunt.loadNpmTasks('grunt-contrib-uglify');

  grunt.initConfig({
    watch: {
      files: [
        'assets/src/js/**/*.js',
        'assets/src/css/sass/**/*.scss'
      ],
      tasks: ['sass', 'concat', 'cssmin', 'uglify']
    },
    sass: {
      dist: {
        options: {
          style: 'expanded',
          sourcemap: 'none'
        },
        files: [{
          expand: true,
          cwd: 'assets/src/css/sass',
          src: ['**/*.scss'],
          dest: 'assets/src/css',
          ext: '.css'
        }]
      }
    },
    concat: {
      css: {
        src: ['node_modules/normalize.css/normalize.css',
          'node_modules/font-awesome/css/font-awesome.css',
          'assets/src/css/main.css'
        ],
        dest: 'assets/src/css/main.css',
      },
    },
    cssmin: {
      target: {
        files: [{
          expand: true,
          cwd: 'assets/src/css/',
          src: ['*.css', '!*.min.css'],
          dest: 'assets/dist/css/',
          ext: '.min.css'
        }]
      }
    },
    uglify: {
      target: {
        files: {
          'assets/dist/js/app.min.js': ['node_modules/jquery/jquery.js',
            'node_modules/perfect-scrollbar/**/perfect-scrollbar.jquery.js',
            'node_modules/showdown/dist/showdown.js',
            'assets/src/js/**/*.js'
          ]
        }
      }
    }
  });

  grunt.registerTask('default', ['sass', 'concat', 'cssmin', 'uglify']);
};