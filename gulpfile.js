/**
 * gulpfile, copy and edit from <https://gist.github.com/squidfunk/120b6f02927fdc9ef9f1>
 */

'use strict';

var gulp = require('gulp');
var child = require('child_process');
var args = require('yargs').argv;
var gulpgo = require('gulp-go');
var mainBowerFiles = require('main-bower-files');
var less = require('gulp-less');
var postcss = require('gulp-postcss');
var sourcemaps = require('gulp-sourcemaps');
var autoprefixer = require('autoprefixer');
var livereload = require('gulp-livereload');
var path = require('path');
var uglify = require('gulp-uglify');
var gulpif = require('gulp-if');
var plumber = require('gulp-plumber');
var util = require('gulp-util');
var notifier = require('node-notifier');
var mincss = require('gulp-minify-css');
var sync = require('gulp-sync')(gulp).sync;

/* ----------------------------------------------------------------------------
 * Locals
 * ------------------------------------------------------------------------- */

/* Application server */
var server = null;

/*
 * Override gulp.src() for nicer error handling.
 */
var src = gulp.src;
gulp.src = function() {
    return src.apply(gulp, arguments)
        .pipe(plumber(function(error) {
            util.log(util.colors.red(
                'Error (' + error.plugin + '): ' + error.message
            ));
            notifier.notify({
                title: 'Error (' + error.plugin + ')',
                message: error.message.split('\n')[0]
            });
            this.emit('end');
        }));
};

var go;

/* ----------------------------------------------------------------------------
 * Assets pipeline
 * ------------------------------------------------------------------------- */

gulp.task('go-run', function() {
    go = gulpgo.run('main.go', [], {
        cwd: __dirname,
        stdio: 'inherit',
        godep: true
    });
});

/*
 * Build bower libs, copy main bower files to ./puglic/lib/
 */
//gulp.task('assets:bower', function() {
    //return gulp.src(mainBowerFiles()).pipe(gulp.dest('./public/lib'));
//});

/*
 * Copy assets files to public
 */
gulp.task('assets:files', function() {
    return gulp.src([
        'assets/files/**/*'
    ]).pipe(gulp.dest('./public/'));
});

/*
 * Build stylesheets from LESS source
 */
gulp.task('assets:stylesheets', function() {
    return gulp.src('assets/css/*.less')
        .pipe(less())
        .pipe(sourcemaps.init())
        .pipe(gulpif(args.production, postcss([
            autoprefixer({
                browsers: ['last 2 versions']
            })
        ])))
        .pipe(gulpif(args.production, mincss()))
        .pipe(sourcemaps.write('.'))
        .pipe(gulp.dest('public/css'))
        .pipe(livereload());
});

/*
 * Build javascripts from source
 */
gulp.task('assets:javascripts', function() {
    return gulp.src('assets/js/*.js')
        .pipe(gulpif(args.production, uglify()))
        .pipe(gulp.dest('public/js'))
        .pipe(livereload());
});

/*
 * Build assets.
 */
gulp.task('assets:build', [
    'assets:stylesheets',
    'assets:javascripts',
    //'assets:bower',
    'assets:files'
]);

gulp.task('assets:watch', function() {
    gulp.watch([
        'assets/css/**/*.less'
    ], ['assets:stylesheets']);

    gulp.watch([
        'assets/js/**/*.js'
    ], ['assets:javascripts']);

    //gulp.watch([
        //'bower.json'
    //], ['assets:bower']);

    gulp.watch([
        'assets/files/**/*'
    ], ['assets:files']);
});

/* ----------------------------------------------------------------------------
 * Application server
 * ------------------------------------------------------------------------- */

/*
 * Build application server.
 */
gulp.task('server:build', function() {
    var build = child.spawnSync('go', ['install']);
    if (build.stderr.length) {
        var lines = build.stderr.toString()
            .split('\n').filter(function(line) {
                return line.length;
            });
        for (var l in lines) {
            util.log(util.colors.red(
                'Error (go install): ' + lines[l]
            ));
        }
        notifier.notify({
            title: 'Error (go install)',
            message: lines
        });
    }
    return build;
});

/*
 * Restart application server.
 */
gulp.task('server:spawn', function() {
    if (server)
        server.kill();

    /* Spawn application server */
    server = child.spawn('main');

    /* Trigger reload upon server start */
    server.stdout.once('data', function() {
        livereload.reload('/');
    });

    /* Pretty print server log output */
    server.stdout.on('data', function(data) {
        var lines = data.toString().split('\n')
        for (var l in lines)
            if (lines[l].length)
                util.log(lines[l]);
    });

    /* Print errors to stdout */
    server.stderr.on('data', function(data) {
        process.stdout.write(data.toString());
    });

    livereload.reload();
});

/*
 * Watch source for changes and restart application server.
 */
gulp.task('server:watch', function() {

    /* Restart application server */
    gulp.watch([
        './tmpls/**/*.tmpl',
        './config.json'
    ], ['server:spawn']);

    /* Rebuild and restart application server */
    gulp.watch([
        './**/*.go',
    ], sync([
        'server:build',
        'server:spawn'
    ], 'server'));
});

/* ----------------------------------------------------------------------------
 * Interface
 * ------------------------------------------------------------------------- */

/*
 * Build assets and application server.
 */
gulp.task('build', [
    'assets:build',
    'server:build'
]);

/*
 * Start asset and server watchdogs and initialize livereload.
 */
gulp.task('watch', [
    'assets:build',
    'server:build'
], function() {
    livereload.listen();
    return gulp.start([
        'assets:watch',
        'server:watch',
        'server:spawn'
    ]);
});

/*
 * Build assets by default.
 */
gulp.task('default', ['build']);
