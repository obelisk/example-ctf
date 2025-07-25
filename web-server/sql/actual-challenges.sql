-- Auto-generated SQL to populate challenges and flags tables
-- Generated from ctf-challenges.json
-- Uses UPSERT operations to update existing challenges

INSERT INTO challenges (id, name, description, category, point_reward_amount, file_asset, text_asset, nested_id) VALUES (
    1,
    'Ready to Mingle',
    'Those people who are ready to mingle must have built this challenge. Good luck.',
    'cryptography',
    1,
    NULL,
    'WJEYAAWOULHWYAPKCAPOPWNPAZPDAJATPKJAOSEHHJKPXAOQYDWSWHGEJPDALWNPUKQNBHWCEOEZEZWOEILHABKNHKKLWJZWHHECKPSWOPDEOHKQOUBHWC',
    1
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    point_reward_amount = EXCLUDED.point_reward_amount,
    file_asset = EXCLUDED.file_asset,
    text_asset = EXCLUDED.text_asset,
    nested_id = EXCLUDED.nested_id;

INSERT INTO flags (challenge_id, flag_value, validation_handler) VALUES (
    1,
    'IDIDASIMPLEFORLOOPANDALLIGOTWASTHISLOUSYFLAG',
    'StringEqual'
)
ON CONFLICT (challenge_id) DO UPDATE SET
    flag_value = EXCLUDED.flag_value,
    validation_handler = EXCLUDED.validation_handler;

INSERT INTO challenges (id, name, description, category, point_reward_amount, file_asset, text_asset, nested_id) VALUES (
    2,
    'Beauty to Behold',
    'Sometimes it''s just nice to have something to look at. Of course, it would be easier if it were decrypted first.',
    'cryptography',
    8,
    'beauty-to-behold.bmp',
    NULL,
    2
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    point_reward_amount = EXCLUDED.point_reward_amount,
    file_asset = EXCLUDED.file_asset,
    text_asset = EXCLUDED.text_asset,
    nested_id = EXCLUDED.nested_id;

INSERT INTO flags (challenge_id, flag_value, validation_handler) VALUES (
    2,
    'icannotbelieveihadtolearnhowtobmpforthis',
    'StringEqual'
)
ON CONFLICT (challenge_id) DO UPDATE SET
    flag_value = EXCLUDED.flag_value,
    validation_handler = EXCLUDED.validation_handler;

INSERT INTO challenges (id, name, description, category, point_reward_amount, file_asset, text_asset, nested_id) VALUES (
    3,
    'A Real Drag',
    'Really how much more to do you want after you''ve come this far? I believe in you to drag yourself across the finish line.',
    'cryptography',
    12,
    NULL,
    'VVROUETVPHBVRNUFVDWEISEMKEEDGSZTZPWASNHVVNFMZBBMRTXUAQTPCPLRTWAJXWWZTALVIVLWVWMAJVBVGAYWECXQMLMPYYSGOMJOKHRKDOVAJEUVUMTWCFHVWSIKIFMLLLKIEBPKPFUGDBYJXHXGWDLYPGIQWDAMIYKYZXSTXSJATRBZLEVXLHIEKQYXASAFZOLMKEVGVUEEWCYADXTJSKGBFLAJIIAVROTNBRILFODGATTLPMFVFIWNJIMLLLAEHYVGAVBFVJIEGMTQWAFHWPHZPZCICLPGOCHDBRRMIIRSRPYFRTSPPKGYXTSOWDEZXUXMZSKSVZFZWJXTKIJIVOURVISHUTNKGJUUYGSSALCBWLATGBGUYHOSKLPOYSWH',
    3
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    point_reward_amount = EXCLUDED.point_reward_amount,
    file_asset = EXCLUDED.file_asset,
    text_asset = EXCLUDED.text_asset,
    nested_id = EXCLUDED.nested_id;

INSERT INTO flags (challenge_id, flag_value, validation_handler) VALUES (
    3,
    'IMTIREDOFCRIBDRAGGINGANDNOWIJUSTWANTTOGOHOMEPLEASE',
    'StringEqual'
)
ON CONFLICT (challenge_id) DO UPDATE SET
    flag_value = EXCLUDED.flag_value,
    validation_handler = EXCLUDED.validation_handler;

INSERT INTO challenges (id, name, description, category, point_reward_amount, file_asset, text_asset, nested_id) VALUES (
    4,
    'Muxing Data Fiveever',
    'There exist algorithms that despite being old and tired we simply cannot fully get rid of. Use this algorithm to solve this challenge.',
    'proof-of-work',
    2,
    NULL,
    'The flag is the preimage containing your work email.',
    1
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    point_reward_amount = EXCLUDED.point_reward_amount,
    file_asset = EXCLUDED.file_asset,
    text_asset = EXCLUDED.text_asset,
    nested_id = EXCLUDED.nested_id;

INSERT INTO flags (challenge_id, flag_value, validation_handler) VALUES (
    4,
    '6',
    'Md5HashOfUsername'
)
ON CONFLICT (challenge_id) DO UPDATE SET
    flag_value = EXCLUDED.flag_value,
    validation_handler = EXCLUDED.validation_handler;

INSERT INTO challenges (id, name, description, category, point_reward_amount, file_asset, text_asset, nested_id) VALUES (
    5,
    'An Annoying Persons Favourite',
    'Sometimes naming software after yourself isn''t as flattering as you''d hope. This software relies on this algorithm despite the fact we should really not be using it anymore.',
    'proof-of-work',
    3,
    NULL,
    'The flag is the preimage containing your work email.',
    2
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    point_reward_amount = EXCLUDED.point_reward_amount,
    file_asset = EXCLUDED.file_asset,
    text_asset = EXCLUDED.text_asset,
    nested_id = EXCLUDED.nested_id;

INSERT INTO flags (challenge_id, flag_value, validation_handler) VALUES (
    5,
    '6',
    'Sha1HashOfUsername'
)
ON CONFLICT (challenge_id) DO UPDATE SET
    flag_value = EXCLUDED.flag_value,
    validation_handler = EXCLUDED.validation_handler;

INSERT INTO challenges (id, name, description, category, point_reward_amount, file_asset, text_asset, nested_id) VALUES (
    6,
    'Buddy Buddy of the Green Lock',
    'The green lock shows that you''re secure (though all those VPN adds on YouTube beg to differ). Where have you seen that lock? And what hashing algorithm is used?',
    'proof-of-work',
    5,
    NULL,
    'The flag is the preimage containing your work email.',
    3
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    point_reward_amount = EXCLUDED.point_reward_amount,
    file_asset = EXCLUDED.file_asset,
    text_asset = EXCLUDED.text_asset,
    nested_id = EXCLUDED.nested_id;

INSERT INTO flags (challenge_id, flag_value, validation_handler) VALUES (
    6,
    '8',
    'Sha256HashOfUsername'
)
ON CONFLICT (challenge_id) DO UPDATE SET
    flag_value = EXCLUDED.flag_value,
    validation_handler = EXCLUDED.validation_handler;

INSERT INTO challenges (id, name, description, category, point_reward_amount, file_asset, text_asset, nested_id) VALUES (
    7,
    'Dont You Dare Call Me Sha3',
    'How dare you you. I''m nothing like my father. I am blockchain forward. Go take a sponge bath and think about what you''ve done.',
    'proof-of-work',
    10,
    NULL,
    'The flag is the preimage containing your work email.',
    4
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    point_reward_amount = EXCLUDED.point_reward_amount,
    file_asset = EXCLUDED.file_asset,
    text_asset = EXCLUDED.text_asset,
    nested_id = EXCLUDED.nested_id;

INSERT INTO flags (challenge_id, flag_value, validation_handler) VALUES (
    7,
    '10',
    'Keccak256HashOfUsername'
)
ON CONFLICT (challenge_id) DO UPDATE SET
    flag_value = EXCLUDED.flag_value,
    validation_handler = EXCLUDED.validation_handler;

INSERT INTO challenges (id, name, description, category, point_reward_amount, file_asset, text_asset, nested_id) VALUES (
    8,
    'The New Web Hotness',
    '',
    'reverse-engineering',
    5,
    'the-new-web-hotness.wasm',
    NULL,
    1
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    point_reward_amount = EXCLUDED.point_reward_amount,
    file_asset = EXCLUDED.file_asset,
    text_asset = EXCLUDED.text_asset,
    nested_id = EXCLUDED.nested_id;

INSERT INTO flags (challenge_id, flag_value, validation_handler) VALUES (
    8,
    'OhWoWiCaNcLeArLySeEtHiSwAsMsTuFfIsThEfUtUrEiHaTeItSoVeRyVeRyLoT',
    'StringEqual'
)
ON CONFLICT (challenge_id) DO UPDATE SET
    flag_value = EXCLUDED.flag_value,
    validation_handler = EXCLUDED.validation_handler;

INSERT INTO challenges (id, name, description, category, point_reward_amount, file_asset, text_asset, nested_id) VALUES (
    9,
    'An Arduous One',
    'This challenge is a bit of a pain, but it is worth it in the end. Good luck.',
    'reverse-engineering',
    12,
    'arduino-build.zip',
    NULL,
    2
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    point_reward_amount = EXCLUDED.point_reward_amount,
    file_asset = EXCLUDED.file_asset,
    text_asset = EXCLUDED.text_asset,
    nested_id = EXCLUDED.nested_id;

INSERT INTO flags (challenge_id, flag_value, validation_handler) VALUES (
    9,
    'HOWWASISUPPOSEDTOKNOWINEEDEDTOBRINGRESISTORS',
    'StringEqual'
)
ON CONFLICT (challenge_id) DO UPDATE SET
    flag_value = EXCLUDED.flag_value,
    validation_handler = EXCLUDED.validation_handler;

INSERT INTO challenges (id, name, description, category, point_reward_amount, file_asset, text_asset, nested_id) VALUES (
    10,
    'Reading Between',
    'Messages hidden in between the messages.',
    'steganography',
    3,
    NULL,
    'The  computing world  has undergone a revolution  since the publication of The C  Programming Language  in  1978. Big computers  are  much  bigger, and personal computers have capabilities that rival mainframes  of a decade ago.  During  this time, C  has changed  too,  although only modestly,  and  it has spread far  beyond its origins as  the language  of the UNIX operating system.

The  growing popularity  of  C,  the changes in the  language over  the years and the creation  of  compilers by groups  not  involved  in  its design, combined  to  demonstrate a  need for a more precise and more  contemporary  definition of the  language than the first  edition  of  this book provided.  In 1983,  the  American National  Standards Institute (ANSI) established a  committee  whose goal was  to  produce  "an  unambiguous  and machine-independent definition  of the language  C",  while still  retaining its spirit.  The result  is the ANSI standard for C. ',
    1
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    point_reward_amount = EXCLUDED.point_reward_amount,
    file_asset = EXCLUDED.file_asset,
    text_asset = EXCLUDED.text_asset,
    nested_id = EXCLUDED.nested_id;

INSERT INTO flags (challenge_id, flag_value, validation_handler) VALUES (
    10,
    'thespacesbeaneasytargetforme',
    'StringEqual'
)
ON CONFLICT (challenge_id) DO UPDATE SET
    flag_value = EXCLUDED.flag_value,
    validation_handler = EXCLUDED.validation_handler;

INSERT INTO challenges (id, name, description, category, point_reward_amount, file_asset, text_asset, nested_id) VALUES (
    11,
    'Looking Deeply',
    'Just seeing this enigmatic graphic is not enough.',
    'steganography',
    2,
    'holland.jpg',
    NULL,
    2
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    point_reward_amount = EXCLUDED.point_reward_amount,
    file_asset = EXCLUDED.file_asset,
    text_asset = EXCLUDED.text_asset,
    nested_id = EXCLUDED.nested_id;

INSERT INTO flags (challenge_id, flag_value, validation_handler) VALUES (
    11,
    'IMNOTSUREIREALLYUNDERSTANDTHEDCTYETTHOUGH',
    'StringEqual'
)
ON CONFLICT (challenge_id) DO UPDATE SET
    flag_value = EXCLUDED.flag_value,
    validation_handler = EXCLUDED.validation_handler;

INSERT INTO challenges (id, name, description, category, point_reward_amount, file_asset, text_asset, nested_id) VALUES (
    12,
    'Listening Intently',
    'Music can be so colourful.',
    'steganography',
    4,
    'listening-intently.mp3',
    NULL,
    3
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    point_reward_amount = EXCLUDED.point_reward_amount,
    file_asset = EXCLUDED.file_asset,
    text_asset = EXCLUDED.text_asset,
    nested_id = EXCLUDED.nested_id;

INSERT INTO flags (challenge_id, flag_value, validation_handler) VALUES (
    12,
    'FULLSENDIT',
    'StringEqual'
)
ON CONFLICT (challenge_id) DO UPDATE SET
    flag_value = EXCLUDED.flag_value,
    validation_handler = EXCLUDED.validation_handler;

INSERT INTO challenges (id, name, description, category, point_reward_amount, file_asset, text_asset, nested_id) VALUES (
    13,
    'Foot Tapping',
    'Something to tap to.',
    'steganography',
    4,
    'foot-tapping.mp3',
    NULL,
    4
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    point_reward_amount = EXCLUDED.point_reward_amount,
    file_asset = EXCLUDED.file_asset,
    text_asset = EXCLUDED.text_asset,
    nested_id = EXCLUDED.nested_id;

INSERT INTO flags (challenge_id, flag_value, validation_handler) VALUES (
    13,
    'ITHOUGHTTHEOTHERAUDIOONEWASWAYBUTMAYBETHATWASJUSTCHATGPT',
    'StringEqual'
)
ON CONFLICT (challenge_id) DO UPDATE SET
    flag_value = EXCLUDED.flag_value,
    validation_handler = EXCLUDED.validation_handler;

INSERT INTO challenges (id, name, description, category, point_reward_amount, file_asset, text_asset, nested_id) VALUES (
    14,
    'Client Side Wasm',
    'Please help us reverse engineer another WASM blob',
    'exam',
    0,
    'client-side-wasm.zip',
    '# Internal Access Issues

We''ve just received this from an agent compromising a tech-company. Looking at network traffic it appears this WASM blob is responsible for generating authentication tokens consumed by a backend. To further their access, we need you to generate a valid JWT where the admin flag is set to true.

Logisitics has provided you with a minimal setup to load the WASM so you can focus on reversing engineering.

Submit the admin JWT as soon as you have it.',
    1
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    point_reward_amount = EXCLUDED.point_reward_amount,
    file_asset = EXCLUDED.file_asset,
    text_asset = EXCLUDED.text_asset,
    nested_id = EXCLUDED.nested_id;

INSERT INTO flags (challenge_id, flag_value, validation_handler) VALUES (
    14,
    '-----BEGIN PUBLIC KEY-----
MCowBQYDK2VwAyEAcg39a8xWjWlwP2Li1R6ep4HRgKx2SIXEsHC4+4ar0FM=
-----END PUBLIC KEY-----',
    'ClientSideWasm'
)
ON CONFLICT (challenge_id) DO UPDATE SET
    flag_value = EXCLUDED.flag_value,
    validation_handler = EXCLUDED.validation_handler;
