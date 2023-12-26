#include <ctime>
#include <immintrin.h>
#include <ncurses.h>
#include <vector>

#ifdef _BENCHMARK
#include <benchmark/benchmark.h>
#endif

double rand_max(double max) {
    return static_cast<double>(rand()) / static_cast<double>(RAND_MAX / max);
}

typedef struct {
    double separation_factor;
    double alignment_factor;
    double cohesion_factor;
    double turn_impulse;
    double margin;
    double separation_threshold;
    double max_speed;
    double min_speed;
} config;

union vec2 {
    struct {
        double y;
        double x;
    } s;
    __m128d m;
};

typedef struct boid {
    vec2 pos;
    vec2 dir;
    std::vector<boid *> flock;
} boid;

double distance(vec2 v1, vec2 v2) {
    // t_low := v1.x - v2.x; t_high := v1.y - v2.y
    auto t = _mm_sub_pd(v1.m, v2.m);
    // t_low := t_low * t_low; t_high := t_high * t_high
    t = _mm_mul_pd(t, t);
    // t_low := t_low + t_high
    t = _mm_hadd_pd(t, t);
    // t_low := sqrt(t_low)
    t = _mm_sqrt_pd(t);
    return _mm_cvtsd_f64(t);
}

class goggle {
  private:
    double r;
    config cfg;

  public:
    std::vector<boid> boids;
    goggle(double, size_t, config, int, int);
    void update(int, int);
};

goggle::goggle(double r, size_t num_boids, config cfg, int max_y, int max_x) {
    for (size_t i = 0; i < num_boids; i++) {
        boid b;
        b.pos.m =
            _mm_set_pd(rand_max(max_y), static_cast<double>(max_x) / 2 -
                                            static_cast<double>(max_y) / 2 +
                                            rand_max(max_y));
        b.dir.m = _mm_set_pd(rand_max(cfg.min_speed), rand_max(cfg.min_speed));
        boids.push_back(b);
    }

    this->r = r;
    this->cfg = cfg;
}

void goggle::update(int max_y, int max_x) {
    for (auto &b : boids) {
        b.flock.clear();
    }

    for (size_t i = 0; i < boids.size(); i++) {
        for (size_t j = i + 1; j < boids.size(); j++) {
            auto &b1 = boids[i];
            auto &b2 = boids[j];
            auto d = distance(b1.pos, b2.pos);
            if (d < this->r) {
                b1.flock.push_back(&b2);
                b2.flock.push_back(&b1);
            }
        }
    }

    for (auto &b : boids) {
        vec2 inv_close = {.m = _mm_setzero_pd()};
        vec2 avg_heading = {.m = _mm_setzero_pd()};
        vec2 avg_pos = {.m = _mm_setzero_pd()};

        for (auto &f : b.flock) {
            if (distance(b.pos, f->pos) < this->cfg.separation_threshold) {
                inv_close.m =
                    _mm_add_pd(inv_close.m, _mm_sub_pd(b.pos.m, f->pos.m));
            }
            avg_heading.m = _mm_add_pd(avg_heading.m, f->dir.m);
            avg_pos.m = _mm_add_pd(avg_pos.m, f->pos.m);
        }

        // inv_close *= separationFactor
        inv_close.m =
            _mm_mul_pd(inv_close.m, _mm_set1_pd(this->cfg.separation_factor));

        auto fs = b.flock.size();
        if (b.flock.size() > 0) {
            // avg_heading /= b.flock.size()
            avg_heading.m = _mm_div_pd(avg_heading.m, _mm_set1_pd(fs));
            avg_heading.m = _mm_mul_pd(avg_heading.m,
                                       _mm_set1_pd(this->cfg.alignment_factor));

            // avg_pos /= b.flock.size()
            avg_pos.m = _mm_div_pd(avg_pos.m, _mm_set1_pd(fs));
            avg_pos.m = _mm_sub_pd(avg_pos.m, b.pos.m);
            avg_pos.m =
                _mm_mul_pd(avg_pos.m, _mm_set1_pd(this->cfg.cohesion_factor));
        }

        b.dir.m = _mm_add_pd(b.dir.m, inv_close.m);
        b.dir.m = _mm_add_pd(b.dir.m, avg_heading.m);
        b.dir.m = _mm_add_pd(b.dir.m, avg_pos.m);

        // change dir by turnImpulse if position is within margin
        b.dir.m = _mm_sub_pd(
            b.dir.m, _mm_and_pd(_mm_set1_pd(this->cfg.turn_impulse),
                                _mm_cmp_pd(_mm_set_pd(max_x - this->cfg.margin,
                                                      max_y - this->cfg.margin),
                                           b.pos.m, _CMP_LT_OQ)));

        // same thing for low x and y
        b.dir.m = _mm_add_pd(
            b.dir.m, _mm_and_pd(_mm_set1_pd(this->cfg.turn_impulse),
                                _mm_cmp_pd(_mm_set1_pd(this->cfg.margin),
                                           b.pos.m, _CMP_GT_OQ)));

        // clamp speed
        auto speed = distance(b.dir, {.m = _mm_setzero_pd()});
        if (speed > this->cfg.max_speed) {
            b.dir.m = _mm_div_pd(b.dir.m, _mm_set1_pd(speed));
            b.dir.m = _mm_mul_pd(b.dir.m, _mm_set1_pd(this->cfg.max_speed));
        } else if (speed < this->cfg.min_speed) {
            b.dir.m = _mm_div_pd(b.dir.m, _mm_set1_pd(speed));
            b.dir.m = _mm_mul_pd(b.dir.m, _mm_set1_pd(this->cfg.min_speed));
        }

        // update position
        b.pos.m = _mm_add_pd(b.pos.m, b.dir.m);
    }
}

#ifdef _BENCHMARK

static void BM_250(benchmark::State &state) {

    config cfg = {.separation_factor = 0.05,
                  .alignment_factor = 0.05,
                  .cohesion_factor = 0.0008,
                  .turn_impulse = 0.2,
                  .margin = 25,
                  .separation_threshold = 2,
                  .max_speed = 2.5,
                  .min_speed = 1};

    goggle g(20, 250, cfg, 1000, 1000);
    for (auto _ : state) {
        g.update(1000, 1000);
    }
}

static void BM_500(benchmark::State &state) {

    config cfg = {.separation_factor = 0.05,
                  .alignment_factor = 0.05,
                  .cohesion_factor = 0.0008,
                  .turn_impulse = 0.2,
                  .margin = 25,
                  .separation_threshold = 2,
                  .max_speed = 2.5,
                  .min_speed = 1};

    goggle g(20, 500, cfg, 1000, 1000);
    for (auto _ : state) {
        g.update(1000, 1000);
    }
}

BENCHMARK(BM_250);
BENCHMARK(BM_500);

BENCHMARK_MAIN();
#else
int main() {
    srand(time(NULL));

    auto win = initscr();
    noecho();
    curs_set(0);
    halfdelay(1);

    int max_y, max_x;
    getmaxyx(win, max_y, max_x);

    config cfg = {.separation_factor = 0.05,
                  .alignment_factor = 0.05,
                  .cohesion_factor = 0.0008,
                  .turn_impulse = 0.2,
                  .margin = 25,
                  .separation_threshold = 2,
                  .max_speed = 2.5,
                  .min_speed = 1};

    goggle g(20, 250, cfg, max_y, max_x);
    g.update(max_y, max_x);

    while (true) {
        getmaxyx(win, max_y, max_x);
        erase();
        for (auto &b : g.boids) {
            mvaddch(b.pos.s.y, b.pos.s.x, 'o');
        }
        g.update(max_y, max_x);

        if (getch() == 'q') {
            break;
        }
    }

    endwin();
    return 0;
}
#endif
