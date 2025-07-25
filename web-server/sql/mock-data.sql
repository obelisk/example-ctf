-- Mock data for CTF database
-- This file contains sample data to simulate the specified user state

-- Clear existing data (optional - uncomment if you want to start fresh)
-- DELETE FROM user_history_log;
-- DELETE FROM user_challenges_completed;
-- DELETE FROM user_aliases;
-- DELETE FROM users;
-- DELETE FROM flags;
-- DELETE FROM challenges;

-- Insert users with their current state
INSERT INTO users (user_email, tokens_available, tokens_burned, points_achieved, exam_challenges_solved, last_exam_challenge_solved_timestamp, last_challenge_solved_timestamp) VALUES
('thanh.nguyen@smartcontract.com', 0, 0, 16, 0, NOW(), NOW() - INTERVAL '2 days 4 hours'),
('javier@smartcontract.com', 0, 0, 11, 0, NOW(), NOW() - INTERVAL '3 days 5 hours 15 minutes'),
('harry.anderson@smartcontract.com', 0, 0, 11, 0, NOW(), NOW() - INTERVAL '1 day 5 hours 45 minutes'),
('maus@smartcontract.com', 0, 0, 8, 0, NOW(), NOW() - INTERVAL '4 days 6 hours 30 minutes'),
('william.spencer@smartcontract.com', 0, 0, 7, 0, NOW(), NOW() - INTERVAL '2 days 7 hours 35 minutes'),
('martin.minkov@smartcontract.com', 0, 0, 1, 0, NOW(), NOW() - INTERVAL '5 hours 10 minutes'),
('sean.hartog@smartcontract.com', 0, 0, 1, 0, NOW(), NOW() - INTERVAL '12 hours 25 minutes'),
('loren.garth@smartcontract.com', 0, 0, 1, 0, NOW(), NOW() - INTERVAL '8 hours 40 minutes'),
('cmc@smartcontract.com', 0, 0, 1, 0, NOW(), NOW() - INTERVAL '6 days 8 hours 55 minutes'),
('monzer@smartcontract.com', 0, 0, 1, 0, NOW(), NOW() - INTERVAL '1 day 9 hours 10 minutes');

-- Sample challenge completions for thanh.nguyen@smartcontract.com (16 points)
-- Based on actual challenges: 1(1pt) + 2(8pt) + 3(12pt) + 4(2pt) + 5(3pt) + 6(5pt) + 7(10pt) + 8(5pt) + 9(12pt) + 10(3pt) + 11(2pt) + 12(4pt) + 13(4pt) = 71 points total
-- For 16 points, completing challenges: 1(1pt) + 4(2pt) + 5(3pt) + 6(5pt) + 10(3pt) + 11(2pt) = 16 points
INSERT INTO user_challenges_completed (user_email, challenge_id, completed_at) VALUES
('thanh.nguyen@smartcontract.com', 1, NOW() - INTERVAL '2 days 6 hours'),
('thanh.nguyen@smartcontract.com', 4, NOW() - INTERVAL '1 day 4 hours 15 minutes'),
('thanh.nguyen@smartcontract.com', 5, NOW() - INTERVAL '1 day 2 hours 30 minutes'),
('thanh.nguyen@smartcontract.com', 6, NOW() - INTERVAL '1 day 1 hour 45 minutes'),
('thanh.nguyen@smartcontract.com', 10, NOW() - INTERVAL '12 hours 30 minutes'),
('thanh.nguyen@smartcontract.com', 11, NOW() - INTERVAL '8 hours 4 minutes');

-- Sample challenge completions for javier@smartcontract.com (11 points)
-- Completing challenges: 1(1pt) + 4(2pt) + 5(3pt) + 6(5pt) = 11 points
INSERT INTO user_challenges_completed (user_email, challenge_id, completed_at) VALUES
('javier@smartcontract.com', 1, NOW() - INTERVAL '3 days 8 hours 30 minutes'),
('javier@smartcontract.com', 4, NOW() - INTERVAL '2 days 6 hours 45 minutes'),
('javier@smartcontract.com', 5, NOW() - INTERVAL '2 days 5 hours 10 minutes'),
('javier@smartcontract.com', 6, NOW() - INTERVAL '1 day 5 hours 15 minutes');

-- Sample challenge completions for harry.anderson@smartcontract.com (11 points)
-- Completing challenges: 1(1pt) + 4(2pt) + 5(3pt) + 6(5pt) = 11 points
INSERT INTO user_challenges_completed (user_email, challenge_id, completed_at) VALUES
('harry.anderson@smartcontract.com', 1, NOW() - INTERVAL '1 day 8 hours 20 minutes'),
('harry.anderson@smartcontract.com', 4, NOW() - INTERVAL '1 day 6 hours 35 minutes'),
('harry.anderson@smartcontract.com', 5, NOW() - INTERVAL '1 day 5 hours 55 minutes'),
('harry.anderson@smartcontract.com', 6, NOW() - INTERVAL '1 day 5 hours 45 minutes');

-- Sample challenge completions for maus@smartcontract.com (8 points)
-- Completing challenges: 1(1pt) + 4(2pt) + 5(3pt) + 10(2pt) = 8 points
INSERT INTO user_challenges_completed (user_email, challenge_id, completed_at) VALUES
('maus@smartcontract.com', 1, NOW() - INTERVAL '4 days 12 hours 40 minutes'),
('maus@smartcontract.com', 4, NOW() - INTERVAL '3 days 10 hours 55 minutes'),
('maus@smartcontract.com', 5, NOW() - INTERVAL '3 days 8 hours 15 minutes'),
('maus@smartcontract.com', 10, NOW() - INTERVAL '2 days 6 hours 30 minutes');

-- Sample challenge completions for william.spencer@smartcontract.com (7 points)
-- Completing challenges: 1(1pt) + 4(2pt) + 5(3pt) + 11(1pt) = 7 points
INSERT INTO user_challenges_completed (user_email, challenge_id, completed_at) VALUES
('william.spencer@smartcontract.com', 1, NOW() - INTERVAL '2 days 10 hours 45 minutes'),
('william.spencer@smartcontract.com', 4, NOW() - INTERVAL '2 days 9 hours'),
('william.spencer@smartcontract.com', 5, NOW() - INTERVAL '1 day 8 hours 20 minutes'),
('william.spencer@smartcontract.com', 11, NOW() - INTERVAL '1 day 7 hours 35 minutes');

-- Sample challenge completions for martin.minkov@smartcontract.com (1 point)
-- Completing challenge: 1(1pt) = 1 point
INSERT INTO user_challenges_completed (user_email, challenge_id, completed_at) VALUES
('martin.minkov@smartcontract.com', 1, NOW() - INTERVAL '5 hours 10 minutes');

-- Sample challenge completions for sean.hartog@smartcontract.com (1 point)
-- Completing challenge: 1(1pt) = 1 point
INSERT INTO user_challenges_completed (user_email, challenge_id, completed_at) VALUES
('sean.hartog@smartcontract.com', 1, NOW() - INTERVAL '12 hours 25 minutes');

-- Sample challenge completions for loren.garth@smartcontract.com (1 point)
-- Completing challenge: 1(1pt) = 1 point
INSERT INTO user_challenges_completed (user_email, challenge_id, completed_at) VALUES
('loren.garth@smartcontract.com', 1, NOW() - INTERVAL '8 hours 40 minutes');

-- Sample challenge completions for cmc@smartcontract.com (1 point)
-- Completing challenge: 1(1pt) = 1 point
INSERT INTO user_challenges_completed (user_email, challenge_id, completed_at) VALUES
('cmc@smartcontract.com', 1, NOW() - INTERVAL '6 days 8 hours 55 minutes');

-- Sample challenge completions for monzer@smartcontract.com (1 point)
-- Completing challenge: 1(1pt) = 1 point
INSERT INTO user_challenges_completed (user_email, challenge_id, completed_at) VALUES
('monzer@smartcontract.com', 1, NOW() - INTERVAL '1 day 9 hours 10 minutes');

-- Insert user history log entries for challenge completions
INSERT INTO user_history_log (user_email, log, date) VALUES
-- thanh.nguyen@smartcontract.com completions
('thanh.nguyen@smartcontract.com', 'Completed challenge 1: added 1 token and 1 points', NOW() - INTERVAL '2 days 6 hours'),
('thanh.nguyen@smartcontract.com', 'Completed challenge 4: added 1 token and 2 points', NOW() - INTERVAL '1 day 4 hours 15 minutes'),
('thanh.nguyen@smartcontract.com', 'Completed challenge 5: added 1 token and 3 points', NOW() - INTERVAL '1 day 2 hours 30 minutes'),
('thanh.nguyen@smartcontract.com', 'Completed challenge 6: added 1 token and 5 points', NOW() - INTERVAL '1 day 1 hour 45 minutes'),
('thanh.nguyen@smartcontract.com', 'Completed challenge 10: added 1 token and 3 points', NOW() - INTERVAL '12 hours 30 minutes'),
('thanh.nguyen@smartcontract.com', 'Completed challenge 11: added 1 token and 2 points', NOW() - INTERVAL '8 hours 4 minutes'),

-- javier completions
('javier@smartcontract.com', 'Completed challenge 1: added 1 token and 1 points', NOW() - INTERVAL '3 days 8 hours 30 minutes'),
('javier@smartcontract.com', 'Completed challenge 4: added 1 token and 2 points', NOW() - INTERVAL '2 days 6 hours 45 minutes'),
('javier@smartcontract.com', 'Completed challenge 5: added 1 token and 3 points', NOW() - INTERVAL '2 days 5 hours 10 minutes'),
('javier@smartcontract.com', 'Completed challenge 6: added 1 token and 5 points', NOW() - INTERVAL '1 day 5 hours 15 minutes'),

-- harry.anderson completions
('harry.anderson@smartcontract.com', 'Completed challenge 1: added 1 token and 1 points', NOW() - INTERVAL '1 day 8 hours 20 minutes'),
('harry.anderson@smartcontract.com', 'Completed challenge 4: added 1 token and 2 points', NOW() - INTERVAL '1 day 6 hours 35 minutes'),
('harry.anderson@smartcontract.com', 'Completed challenge 5: added 1 token and 3 points', NOW() - INTERVAL '1 day 5 hours 55 minutes'),
('harry.anderson@smartcontract.com', 'Completed challenge 6: added 1 token and 5 points', NOW() - INTERVAL '1 day 5 hours 45 minutes'),

-- maus completions
('maus@smartcontract.com', 'Completed challenge 1: added 1 token and 1 points', NOW() - INTERVAL '4 days 12 hours 40 minutes'),
('maus@smartcontract.com', 'Completed challenge 4: added 1 token and 2 points', NOW() - INTERVAL '3 days 10 hours 55 minutes'),
('maus@smartcontract.com', 'Completed challenge 5: added 1 token and 3 points', NOW() - INTERVAL '3 days 8 hours 15 minutes'),
('maus@smartcontract.com', 'Completed challenge 10: added 1 token and 2 points', NOW() - INTERVAL '2 days 6 hours 30 minutes'),

-- william.spencer completions
('william.spencer@smartcontract.com', 'Completed challenge 1: added 1 token and 1 points', NOW() - INTERVAL '2 days 10 hours 45 minutes'),
('william.spencer@smartcontract.com', 'Completed challenge 4: added 1 token and 2 points', NOW() - INTERVAL '2 days 9 hours'),
('william.spencer@smartcontract.com', 'Completed challenge 5: added 1 token and 3 points', NOW() - INTERVAL '1 day 8 hours 20 minutes'),
('william.spencer@smartcontract.com', 'Completed challenge 11: added 1 token and 2 points', NOW() - INTERVAL '1 day 7 hours 35 minutes'),

-- Single challenge completions
('martin.minkov@smartcontract.com', 'Completed challenge 1: added 1 token and 1 points', NOW() - INTERVAL '5 hours 10 minutes'),
('sean.hartog@smartcontract.com', 'Completed challenge 1: added 1 token and 1 points', NOW() - INTERVAL '12 hours 25 minutes'),
('loren.garth@smartcontract.com', 'Completed challenge 1: added 1 token and 1 points', NOW() - INTERVAL '8 hours 40 minutes'),
('cmc@smartcontract.com', 'Completed challenge 1: added 1 token and 1 points', NOW() - INTERVAL '6 days 8 hours 55 minutes'),
('monzer@smartcontract.com', 'Completed challenge 1: added 1 token and 1 points', NOW() - INTERVAL '1 day 9 hours 10 minutes');

-- Insert some sample aliases for users
INSERT INTO user_aliases (user_email, alias, created_at) VALUES
('thanh.nguyen@smartcontract.com', 'timweri', NOW() - INTERVAL '1 day');

-- Add many wrong flag attempts for cmc@smartcontract.com, maus@smartcontract.com, and javier@smartcontract.com
INSERT INTO user_history_log (user_email, log, date) VALUES
-- cmc@smartcontract.com wrong attempts (many attempts across different challenges)
('cmc@smartcontract.com', 'Wrong flag attempt for challenge 2: beauty_wrong1', NOW() - INTERVAL '6 days 12 hours'),
('cmc@smartcontract.com', 'Wrong flag attempt for challenge 2: beauty_wrong2', NOW() - INTERVAL '6 days 11 hours 45 minutes'),
('cmc@smartcontract.com', 'Wrong flag attempt for challenge 2: beauty_wrong3', NOW() - INTERVAL '6 days 11 hours 30 minutes'),
('cmc@smartcontract.com', 'Wrong flag attempt for challenge 2: beauty_wrong4', NOW() - INTERVAL '6 days 11 hours 15 minutes'),
('cmc@smartcontract.com', 'Wrong flag attempt for challenge 2: beauty_wrong5', NOW() - INTERVAL '6 days 11 hours'),
('cmc@smartcontract.com', 'Wrong flag attempt for challenge 3: drag_wrong1', NOW() - INTERVAL '6 days 10 hours 30 minutes'),
('cmc@smartcontract.com', 'Wrong flag attempt for challenge 3: drag_wrong2', NOW() - INTERVAL '6 days 10 hours 15 minutes'),
('cmc@smartcontract.com', 'Wrong flag attempt for challenge 3: drag_wrong3', NOW() - INTERVAL '6 days 10 hours'),
('cmc@smartcontract.com', 'Wrong flag attempt for challenge 4: muxing_wrong1', NOW() - INTERVAL '6 days 9 hours 45 minutes'),
('cmc@smartcontract.com', 'Wrong flag attempt for challenge 4: muxing_wrong2', NOW() - INTERVAL '6 days 9 hours 30 minutes'),
('cmc@smartcontract.com', 'Wrong flag attempt for challenge 5: annoying_wrong1', NOW() - INTERVAL '6 days 9 hours 15 minutes'),
('cmc@smartcontract.com', 'Wrong flag attempt for challenge 5: annoying_wrong2', NOW() - INTERVAL '6 days 9 hours'),
('cmc@smartcontract.com', 'Wrong flag attempt for challenge 6: buddy_wrong1', NOW() - INTERVAL '6 days 8 hours 45 minutes'),
('cmc@smartcontract.com', 'Wrong flag attempt for challenge 6: buddy_wrong2', NOW() - INTERVAL '6 days 8 hours 30 minutes'),
('cmc@smartcontract.com', 'Wrong flag attempt for challenge 7: sha3_wrong1', NOW() - INTERVAL '6 days 8 hours 15 minutes'),
('cmc@smartcontract.com', 'Wrong flag attempt for challenge 8: wasm_wrong1', NOW() - INTERVAL '6 days 8 hours'),
('cmc@smartcontract.com', 'Wrong flag attempt for challenge 8: wasm_wrong2', NOW() - INTERVAL '6 days 7 hours 45 minutes'),
('cmc@smartcontract.com', 'Wrong flag attempt for challenge 9: arduino_wrong1', NOW() - INTERVAL '6 days 7 hours 30 minutes'),
('cmc@smartcontract.com', 'Wrong flag attempt for challenge 10: reading_wrong1', NOW() - INTERVAL '6 days 7 hours 15 minutes'),
('cmc@smartcontract.com', 'Wrong flag attempt for challenge 10: reading_wrong2', NOW() - INTERVAL '6 days 7 hours'),

-- maus@smartcontract.com wrong attempts (many attempts across different challenges)
('maus@smartcontract.com', 'Wrong flag attempt for challenge 2: beauty_wrong1', NOW() - INTERVAL '4 days 15 hours'),
('maus@smartcontract.com', 'Wrong flag attempt for challenge 2: beauty_wrong2', NOW() - INTERVAL '4 days 14 hours 45 minutes'),
('maus@smartcontract.com', 'Wrong flag attempt for challenge 2: beauty_wrong3', NOW() - INTERVAL '4 days 14 hours 30 minutes'),
('maus@smartcontract.com', 'Wrong flag attempt for challenge 2: beauty_wrong4', NOW() - INTERVAL '4 days 14 hours 15 minutes'),
('maus@smartcontract.com', 'Wrong flag attempt for challenge 2: beauty_wrong5', NOW() - INTERVAL '4 days 14 hours'),
('maus@smartcontract.com', 'Wrong flag attempt for challenge 2: beauty_wrong6', NOW() - INTERVAL '4 days 13 hours 45 minutes'),
('maus@smartcontract.com', 'Wrong flag attempt for challenge 3: drag_wrong1', NOW() - INTERVAL '4 days 13 hours 30 minutes'),
('maus@smartcontract.com', 'Wrong flag attempt for challenge 3: drag_wrong2', NOW() - INTERVAL '4 days 13 hours 15 minutes'),
('maus@smartcontract.com', 'Wrong flag attempt for challenge 3: drag_wrong3', NOW() - INTERVAL '4 days 13 hours'),
('maus@smartcontract.com', 'Wrong flag attempt for challenge 3: drag_wrong4', NOW() - INTERVAL '4 days 12 hours 45 minutes'),
('maus@smartcontract.com', 'Wrong flag attempt for challenge 4: muxing_wrong1', NOW() - INTERVAL '4 days 12 hours 30 minutes'),
('maus@smartcontract.com', 'Wrong flag attempt for challenge 4: muxing_wrong2', NOW() - INTERVAL '4 days 12 hours 15 minutes'),
('maus@smartcontract.com', 'Wrong flag attempt for challenge 5: annoying_wrong1', NOW() - INTERVAL '4 days 12 hours'),
('maus@smartcontract.com', 'Wrong flag attempt for challenge 5: annoying_wrong2', NOW() - INTERVAL '4 days 11 hours 45 minutes'),
('maus@smartcontract.com', 'Wrong flag attempt for challenge 6: buddy_wrong1', NOW() - INTERVAL '4 days 11 hours 30 minutes'),
('maus@smartcontract.com', 'Wrong flag attempt for challenge 6: buddy_wrong2', NOW() - INTERVAL '4 days 11 hours 15 minutes'),
('maus@smartcontract.com', 'Wrong flag attempt for challenge 7: sha3_wrong1', NOW() - INTERVAL '4 days 11 hours'),
('maus@smartcontract.com', 'Wrong flag attempt for challenge 8: wasm_wrong1', NOW() - INTERVAL '4 days 10 hours 45 minutes'),
('maus@smartcontract.com', 'Wrong flag attempt for challenge 9: arduino_wrong1', NOW() - INTERVAL '4 days 10 hours 30 minutes'),
('maus@smartcontract.com', 'Wrong flag attempt for challenge 10: reading_wrong1', NOW() - INTERVAL '4 days 10 hours 15 minutes'),
('maus@smartcontract.com', 'Wrong flag attempt for challenge 11: looking_wrong1', NOW() - INTERVAL '4 days 10 hours'),
('maus@smartcontract.com', 'Wrong flag attempt for challenge 12: listening_wrong1', NOW() - INTERVAL '4 days 9 hours 45 minutes'),
('maus@smartcontract.com', 'Wrong flag attempt for challenge 13: foot_wrong1', NOW() - INTERVAL '4 days 9 hours 30 minutes'),

-- javier@smartcontract.com wrong attempts (many attempts across different challenges)
('javier@smartcontract.com', 'Wrong flag attempt for challenge 2: beauty_wrong1', NOW() - INTERVAL '3 days 10 hours'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 2: beauty_wrong2', NOW() - INTERVAL '3 days 9 hours 45 minutes'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 2: beauty_wrong3', NOW() - INTERVAL '3 days 9 hours 30 minutes'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 2: beauty_wrong4', NOW() - INTERVAL '3 days 9 hours 15 minutes'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 2: beauty_wrong5', NOW() - INTERVAL '3 days 9 hours'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 2: beauty_wrong6', NOW() - INTERVAL '3 days 8 hours 45 minutes'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 2: beauty_wrong7', NOW() - INTERVAL '3 days 8 hours 30 minutes'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 2: beauty_wrong8', NOW() - INTERVAL '3 days 8 hours 15 minutes'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 3: drag_wrong1', NOW() - INTERVAL '3 days 8 hours'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 3: drag_wrong2', NOW() - INTERVAL '3 days 7 hours 45 minutes'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 3: drag_wrong3', NOW() - INTERVAL '3 days 7 hours 30 minutes'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 3: drag_wrong4', NOW() - INTERVAL '3 days 7 hours 15 minutes'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 3: drag_wrong5', NOW() - INTERVAL '3 days 7 hours'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 4: muxing_wrong1', NOW() - INTERVAL '3 days 6 hours 45 minutes'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 4: muxing_wrong2', NOW() - INTERVAL '3 days 6 hours 30 minutes'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 4: muxing_wrong3', NOW() - INTERVAL '3 days 6 hours 15 minutes'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 5: annoying_wrong1', NOW() - INTERVAL '3 days 6 hours'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 5: annoying_wrong2', NOW() - INTERVAL '3 days 5 hours 45 minutes'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 6: buddy_wrong1', NOW() - INTERVAL '3 days 5 hours 30 minutes'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 6: buddy_wrong2', NOW() - INTERVAL '3 days 5 hours 15 minutes'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 7: sha3_wrong1', NOW() - INTERVAL '3 days 5 hours'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 8: wasm_wrong1', NOW() - INTERVAL '3 days 4 hours 45 minutes'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 8: wasm_wrong2', NOW() - INTERVAL '3 days 4 hours 30 minutes'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 9: arduino_wrong1', NOW() - INTERVAL '3 days 4 hours 15 minutes'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 9: arduino_wrong2', NOW() - INTERVAL '3 days 4 hours'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 10: reading_wrong1', NOW() - INTERVAL '3 days 3 hours 45 minutes'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 10: reading_wrong2', NOW() - INTERVAL '3 days 3 hours 30 minutes'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 11: looking_wrong1', NOW() - INTERVAL '3 days 3 hours 15 minutes'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 12: listening_wrong1', NOW() - INTERVAL '3 days 3 hours'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 13: foot_wrong1', NOW() - INTERVAL '3 days 2 hours 45 minutes'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 14: wasm_exam_wrong1', NOW() - INTERVAL '3 days 2 hours 30 minutes'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 14: wasm_exam_wrong2', NOW() - INTERVAL '3 days 2 hours 15 minutes'),
('javier@smartcontract.com', 'Wrong flag attempt for challenge 14: wasm_exam_wrong3', NOW() - INTERVAL '3 days 2 hours'),

-- Add some wrong flag attempts for other users for realism
('thanh.nguyen@smartcontract.com', 'Wrong flag attempt for challenge 2: wrong_attempt', NOW() - INTERVAL '2 days 8 hours'),
('harry.anderson@smartcontract.com', 'Wrong flag attempt for challenge 2: another_wrong', NOW() - INTERVAL '1 day 10 hours'),
('william.spencer@smartcontract.com', 'Wrong flag attempt for challenge 2: incorrect_flag', NOW() - INTERVAL '2 days 12 hours'),
('martin.minkov@smartcontract.com', 'Wrong flag attempt for challenge 2: not_right', NOW() - INTERVAL '6 hours'),
('sean.hartog@smartcontract.com', 'Wrong flag attempt for challenge 2: try_again', NOW() - INTERVAL '13 hours'); 