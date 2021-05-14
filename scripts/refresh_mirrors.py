import filecmp
import logging
import os
import sys
import tempfile
import time
import yaml
from prettytable import PrettyTable

LOG_FORMAT = "%(asctime)s - %(levelname)s - %(message)s"
logging.basicConfig(level=logging.INFO, format=LOG_FORMAT)


def judge_statement(command):
    if os.system(command) != 0:
        error_string = command + '  failed :('
        logging.error(error_string)
        sys.exit(1)


def init_mirrors():
    logging.info('start to init mirrors source info')
    if fork_repo not in os.listdir(os.getcwd()):
        logging.error('[init] Error! directory {} does not exists. Check out whether failed in git clone.'.format(fork_repo))
        sys.exit(1)
    current_mirrors = []
    repo_mirrors = []
    judge_statement('mirrorbits list > mirrors.txt')
    with open('mirrors.txt', 'r') as f:
        for line in f.readlines():
            current_mirrors.append(line.split()[0])
    current_mirrors = [(x + '.yaml') for x in current_mirrors[1:]]
    judge_statement('rm mirrors.txt')
    for i in os.listdir('{}/{}'.format(fork_repo, mirrors_dir)):
        if i.endswith('.yaml'):
            repo_mirrors.append(i)
    for i in current_mirrors:
        if i not in repo_mirrors:
            judge_statement('yes | mirrorbits remove {}'.format(i[:-5]))
            logging.info('[init] remove mirror: {}'.format(i[:-5]))
    for i in repo_mirrors:
        if i not in current_mirrors:
            f = open(os.path.join(fork_repo, mirrors_dir, i), 'r')
            mirror_info = yaml.load(f.read(), Loader=yaml.Loader)
            f.close()
            mirror_name = mirror_info['Name']
            if i[:-5] != mirror_name:
                logging.error('Error! {}: filename does not match the name of the mirror'.format(i[:-5]))
                sys.exit(1)
            try:
                admin_email = mirror_info['AdminEmail']
                admin_name = mirror_info['AdminName']
                as_only = mirror_info['ASOnly']
                continent_only = mirror_info['ContinentOnly']
                country_only = mirror_info['CountryOnly']
                ftp_url = mirror_info['FtpURL']
                http_url = mirror_info['HttpURL']
                rsync_url = mirror_info['RsyncURL']
                score = mirror_info['Score']
                sponsor_logo = mirror_info['SponsorLogoURL']
                sponsor_name = mirror_info['SponsorName']
                sponsor_url = mirror_info['SponsorURL']
                command_string = 'mirrorbits add -admin-email="{0}" -admin-name="{1}" -as-only="{2}" ' \
                                 '-continent-only="{3}" -country-only="{4}" -ftp="{5}" -http="{6}" -rsync="{7}" ' \
                                 '-score="{8}" -sponsor-logo="{9}" -sponsor-name="{10}" -sponsor-url="{11}" {12}'.format(
                    admin_email, admin_name, as_only, continent_only, country_only, ftp_url, http_url,
                    rsync_url, score, sponsor_logo, sponsor_name, sponsor_url, mirror_name)
                judge_statement(command_string)
                judge_statement('mirrorbits enable {}'.format(mirror_name))
                pt = PrettyTable(['Key', 'Value'])
                logging.info('[init] add a new mirror: {}, details are below'.format(i[:-5]))
                pt.add_row(['Name', i[:-5]])
                pt.add_row(['AdminEmail', admin_email])
                pt.add_row(['AdminName', admin_name])
                pt.add_row(['ASOnly', as_only])
                pt.add_row(['ContinentOnly', continent_only])
                pt.add_row(['CountryOnly', country_only])
                pt.add_row(['FtpURL', ftp_url])
                pt.add_row(['HttpURL', http_url])
                pt.add_row(['RsyncURL', rsync_url])
                pt.add_row(['Score', score])
                pt.add_row(['SponsorLogoURL', sponsor_logo])
                pt.add_row(['SponsorName', sponsor_name])
                pt.add_row(['SponsorURL', sponsor_url])
                logging.info('\n' + str(pt))
            except KeyError as e:
                logging.error(e)
                exit(1)
        else:
            judge_statement('mirrorbits edit -mirror-file {} {}'.format(os.path.abspath(os.path.join(fork_repo, mirrors_dir, i)), i[:-5]))
            logging.info('[init] update mirror: {}'.format(i[:-5]))


def sync_and_refresh():
    temp_dir = tempfile.gettempdir()
    before_yaml_lst = []
    yaml_lst = []
    # before sync
    logging.info('before sync')
    if mirrors_dir not in os.listdir(fork_repo):
        logging.error('Error! mirrors dir does not exists, exit...')
        sys.exit(1)
    for i in os.listdir('{}/{}'.format(fork_repo, mirrors_dir)):
        if i.endswith('.yaml'):
            before_yaml_lst.append(i)
            judge_statement('cp {} {}'.format(os.path.join(fork_repo, mirrors_dir, i), os.path.join(temp_dir, i)))
    # git sync
    logging.info('git sync')
    judge_statement('cd {}; yes | git sync'.format(fork_repo))
    # after sync
    logging.info('after sync')
    if mirrors_dir not in os.listdir(fork_repo):
        logging.error('Error! mirrors dir does not exists after git sync, exit...')
        sys.exit(1)
    for i in os.listdir('{}/{}'.format(fork_repo, mirrors_dir)):
        if i.endswith('.yaml'):
            yaml_lst.append(i)
    # update mirrors info
    for i in before_yaml_lst:
        if i not in yaml_lst:
            judge_statement('yes | mirrorbits remove {}'.format(i[:-5]))
            logging.info('remove mirror: {}'.format(i[:-5]))
    for i in yaml_lst:
        if i not in before_yaml_lst:
            f = open(os.path.join(fork_repo, mirrors_dir, i), 'r')
            mirror_info = yaml.load(f.read(), Loader=yaml.Loader)
            f.close()
            mirror_name = mirror_info['Name']
            if i[:-5] != mirror_name:
                logging.error('Error! {}: filename does not match the name of the mirror'.format(i[:-5]))
                sys.exit(1)
            try:
                admin_email = mirror_info['AdminEmail']
                admin_name = mirror_info['AdminName']
                as_only = mirror_info['ASOnly']
                continent_only = mirror_info['ContinentOnly']
                country_only = mirror_info['CountryOnly']
                ftp_url = mirror_info['FtpURL']
                http_url = mirror_info['HttpURL']
                rsync_url = mirror_info['RsyncURL']
                score = mirror_info['Score']
                sponsor_logo = mirror_info['SponsorLogoURL']
                sponsor_name = mirror_info['SponsorName']
                sponsor_url = mirror_info['SponsorURL']
                command_string = 'mirrorbits add -admin-email="{0}" -admin-name="{1}" -as-only="{2}" ' \
                                 '-continent-only="{3}" -country-only="{4}" -ftp="{5}" -http="{6}" -rsync="{7}" ' \
                                 '-score="{8}" -sponsor-logo="{9}" -sponsor-name="{10}" -sponsor-url="{11}" {12}'.format(
                    admin_email, admin_name, as_only, continent_only, country_only, ftp_url, http_url,
                    rsync_url, score, sponsor_logo, sponsor_name, sponsor_url, mirror_name)
                judge_statement(command_string)
                judge_statement('mirrorbits enable {}'.format(mirror_name))
                pt = PrettyTable(['Key', 'Value'])
                logging.info('add a new mirror: {}, details are below'.format(i[:-5]))
                pt.add_row(['Name', i[:-5]])
                pt.add_row(['AdminEmail', admin_email])
                pt.add_row(['AdminName', admin_name])
                pt.add_row(['ASOnly', as_only])
                pt.add_row(['ContinentOnly', continent_only])
                pt.add_row(['CountryOnly', country_only])
                pt.add_row(['FtpURL', ftp_url])
                pt.add_row(['HttpURL', http_url])
                pt.add_row(['RsyncURL', rsync_url])
                pt.add_row(['Score', score])
                pt.add_row(['SponsorLogoURL', sponsor_logo])
                pt.add_row(['SponsorName', sponsor_name])
                pt.add_row(['SponsorURL', sponsor_url])
                logging.info('\n' + str(pt))
            except KeyError as e:
                logging.error(e)
                exit(1)
        else:
            if filecmp.cmp(os.path.join(fork_repo, mirrors_dir, i), os.path.join(temp_dir, i), shallow=True):
                continue
            else:
                judge_statement('mirrorbits edit -mirror-file {} {}'.format(os.path.abspath(os.path.join(fork_repo, mirrors_dir, i)), i[:-5]))
                logging.info('update mirror: {}'.format(i[:-5]))
    # clean temp files
    for i in before_yaml_lst:
        judge_statement('rm {}'.format(os.path.join(temp_dir, i)))
        logging.info('remove temp file {}'.format(os.path.join(temp_dir, i)))
    time.sleep(sleep_time)


if __name__ == '__main__':
    with open('refresh_mirrors.yaml', 'r') as fp:
        repo_info = yaml.load(fp.read(), Loader=yaml.Loader)
    fork_url = repo_info['fork_url']
    fork_repo = fork_url.split('/')[-1].split('.')[0]
    mirrors_dir = repo_info['mirrors_dir']
    sleep_time = repo_info['sleep_time']
    if fork_repo in os.listdir(os.getcwd()):
        judge_statement('rm -rf {}'.format(fork_repo))
    # get remote repo code
    logging.info('get remote repo code')
    judge_statement('git clone {}'.format(fork_url))
    # init mirrors source info
    init_mirrors()
    while True:
        sync_and_refresh()
