INSERT INTO users (id, name, email) VALUES
('u-ada', 'Ada Prasetya', 'ada@example.test'),
('u-bima', 'Bima Santoso', 'bima@example.test'),
('u-citra', 'Citra Putri', 'citra@example.test'),
('u-danu', 'Danu Wijaya', 'danu@example.test'),
('u-eka', 'Eka Sari', 'eka@example.test'),
('u-farah', 'Farah Rahman', 'farah@example.test');

INSERT INTO rooms (id, name, capacity, amenities, description) VALUES
('r-garden', 'Garden Room', 4, '["Whiteboard","Video display"]', 'A quiet room beside the courtyard.'),
('r-harbor', 'Harbor Room', 6, '["Whiteboard","Conference phone"]', 'A flexible meeting space near the library.'),
('r-lantern', 'Lantern Room', 6, '["Video display","Speakers"]', 'A bright room for group discussions.'),
('r-maple', 'Maple Room', 10, '["Whiteboard","Video display","Speakers"]', 'A larger room for workshops.'),
('r-orbit', 'Orbit Room', 2, '["Whiteboard"]', 'A compact space for one-to-one meetings.');

INSERT INTO bookings (id, room_id, organizer_id, starts_at, ends_at, purpose) VALUES
('b-101', 'r-garden', 'u-ada', '2030-05-13 09:00+07', '2030-05-13 10:00+07', 'Study group'),
('b-102', 'r-harbor', 'u-bima', '2030-05-13 10:00+07', '2030-05-13 11:30+07', 'Club planning'),
('b-103', 'r-lantern', 'u-citra', '2030-05-13 13:00+07', '2030-05-13 14:00+07', 'Project review'),
('b-104', 'r-maple', 'u-danu', '2030-05-14 09:30+07', '2030-05-14 11:00+07', 'Workshop rehearsal'),
('b-105', 'r-orbit', 'u-eka', '2030-05-14 15:00+07', '2030-05-14 16:00+07', 'Mentoring'),
('b-106', 'r-harbor', 'u-farah', '2030-05-15 13:00+07', '2030-05-15 14:30+07', 'Editorial meeting');

INSERT INTO calendar_events (id, user_id, title, starts_at, ends_at) VALUES
('e-101', 'u-ada', 'Lab class', '2030-05-13 10:00+07', '2030-05-13 12:00+07'),
('e-102', 'u-bima', 'Lecture', '2030-05-13 09:00+07', '2030-05-13 10:00+07'),
('e-103', 'u-citra', 'Committee check-in', '2030-05-13 11:00+07', '2030-05-13 12:00+07'),
('e-104', 'u-danu', 'Office hours', '2030-05-13 14:00+07', '2030-05-13 15:30+07'),
('e-105', 'u-eka', 'Seminar', '2030-05-14 10:00+07', '2030-05-14 12:00+07'),
('e-106', 'u-farah', 'Team catch-up', '2030-05-14 13:00+07', '2030-05-14 14:00+07'),
('e-107', 'u-ada', 'Volunteer shift', '2030-05-15 09:00+07', '2030-05-15 10:30+07'),
('e-108', 'u-bima', 'Design review', '2030-05-15 11:00+07', '2030-05-15 12:00+07');
