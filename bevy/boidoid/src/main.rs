use bevy::prelude::*;
use rand::{thread_rng, Rng};

const NUM_BOIDS: usize = 1500;
const RADIUS: f32 = 60.;
const SEPARATION_FACTOR: f32 = 0.1;
const ALIGNMENT_FACTOR: f32 = 0.05;
const COHESION_FACTOR: f32 = 0.0008;
const TURN_IMPULSE: f32 = 0.1;
const MARGIN: f32 = 300.;
const SEPARATION_THRESHOLD: f32 = 10.;
const MAX_SPEED: f32 = 3.;
const MIN_SPEED: f32 = 1.;

#[derive(Component, Clone)]
struct Boidoid {
    direction: Vec3,
}

fn main() {
    App::new()
        .add_plugins(DefaultPlugins)
        .add_systems(Startup, setup)
        .add_systems(Update, update_boids)
        .run();
}

fn setup(mut commands: Commands, window: Query<&Window>) {
    commands.spawn(Camera2dBundle::default());

    let window = window.single();
    let width = window.width() / 2.;
    let height = window.height() / 2.;

    let mut rng = thread_rng();

    for _ in 0..NUM_BOIDS {
        commands.spawn((
            SpriteBundle {
                transform: Transform {
                    translation: Vec3::new(
                        rng.gen_range(-width..width),
                        rng.gen_range(-height..height),
                        0.,
                    ),
                    scale: Vec3::new(1., 1., 0.),
                    ..default()
                },
                sprite: Sprite {
                    color: Color::WHITE,
                    custom_size: Some(Vec2 { x: 5., y: 5. }),
                    ..default()
                },
                ..default()
            },
            Boidoid {
                direction: Vec3::new(rng.gen_range(-1f32..1f32), rng.gen_range(-1f32..1f32), 0.),
            },
        ));
    }
}

fn update_boids(
    time: Res<Time>,
    window: Query<&Window>,
    mut boids: Query<(&mut Boidoid, &mut Transform)>,
) {
    let window = window.single();
    let width = window.width();
    let height = window.height();

    let mut new_directions = Vec::new();

    for (i, b) in boids.iter().enumerate() {
        let mut dir = b.0.direction;
        let pos = b.1.translation;
        let mut flock = Vec::with_capacity(NUM_BOIDS);

        for (j, b1) in boids.iter().enumerate() {
            if i != j {
                if pos.distance(b1.1.translation) < RADIUS {
                    flock.push(b1);
                }
            }
        }

        let mut inv_close = Vec3::ZERO;
        let mut avg_heading = Vec3::ZERO;
        let mut avg_position = Vec3::ZERO;

        flock.iter().for_each(|f| {
            if pos.distance(f.1.translation) < SEPARATION_THRESHOLD {
                inv_close += pos - f.1.translation;
            }

            avg_position += f.1.translation;
            avg_heading += f.0.direction;
        });

        inv_close *= SEPARATION_FACTOR;

        if flock.len() > 0 {
            avg_heading /= flock.len() as f32;
            avg_heading *= ALIGNMENT_FACTOR;

            avg_position /= flock.len() as f32;
            avg_position -= pos;
            avg_position *= COHESION_FACTOR;
        }

        dir += avg_heading;
        dir += avg_position;
        dir += inv_close;

        if pos.x < -(width / 2.) + MARGIN {
            dir.x += TURN_IMPULSE;
        }
        if pos.x > (width / 2.) - MARGIN {
            dir.x -= TURN_IMPULSE;
        }
        if pos.y < -(height / 2.) + MARGIN {
            dir.y += TURN_IMPULSE;
        }
        if pos.y > (height / 2.) - MARGIN {
            dir.y -= TURN_IMPULSE;
        }

        let speed = dir.distance(Vec3::ZERO);
        if speed > MAX_SPEED {
            dir /= speed;
            dir *= MAX_SPEED;
        } else if speed < MIN_SPEED {
            dir /= speed;
            dir *= MIN_SPEED;
        }

        new_directions.push(dir);
    }

    let mut i = 0;
    for mut b in &mut boids {
        let new_dir = new_directions[i] * 60. * time.delta_seconds();
        b.0.direction = new_dir;
        b.1.translation += new_dir;
        i += 1;
    }
}
