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


def sync_and_refresh(n):
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
            if os.system('cp {} {}'.format(os.path.join(fork_repo, mirrors_dir, i), os.path.join(temp_dir, i))) != 0:
                logging.error('copy temp file failed :(')
                sys.exit(1)
    # git sync
    logging.info('git sync')
    if os.system('cd {}; yes | git sync'.format(fork_repo)) != 0:
        logging.error('git sync failed :(')
        sys.exit(1)
    # after sync
    logging.info('after sync')
    if mirrors_dir not in os.listdir(fork_repo):
        logging.error('Error! mirrors dir does not exists after git sync, exit...')
        sys.exit(1)
    for i in os.listdir('{}/{}'.format(fork_repo, mirrors_dir)):
        if i.endswith('.yaml'):
            yaml_lst.append(i)
    # update mirrors info
    for i in yaml_lst:
        if i not in before_yaml_lst:
            f = open(os.path.join(fork_repo, mirrors_dir, i), 'r')
            mirror_info = yaml.load(f.read(), Loader=yaml.Loader)
            f.close()
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
                if os.system(
                    'mirrorbits add --name "{0}" -admin-email "{1}" -admin-name "{2}" -as-only "{3}" '
                    '-continent-only "{4}" -country-only "{5}" -ftp "{6}" -http "{7}" -rsync "{8}" -score "{9}" '
                    '-sponsor-logo "{10}" -sponsor-name "{11}" -sponsor-url "{12}"'.format(
                        i[:-5], admin_email, admin_name, as_only, continent_only, country_only, ftp_url, http_url,
                        rsync_url, score, sponsor_logo, sponsor_name, sponsor_url)) != 0:
                    logging.error('mirrorbits add failed :(')
                    sys.exit(1)
                pt = PrettyTable(['Key', 'Value'])
                logging.info('add a new mirror: {}, details are below'.format(i[:-5]))
                pt.add_row('Name: {}'.format(i[:-5]))
                pt.add_row('AdminEmail: {}'.format(admin_email))
                pt.add_row('AdminName: {}'.format(admin_name))
                pt.add_row('ASOnly: {}'.format(as_only))
                pt.add_row('ContinentOnly: {}'.format(continent_only))
                pt.add_row('CountryOnly: {}'.format(country_only))
                pt.add_row('FtpURL: {}'.format(ftp_url))
                pt.add_row('HttpURL: {}'.format(http_url))
                pt.add_row('RsyncURL: {}'.format(rsync_url))
                pt.add_row('Score: {}'.format(score))
                pt.add_row('SponsorLogoURL: {}'.format(sponsor_logo))
                pt.add_row('SponsorName: {}'.format(sponsor_name))
                pt.add_row('SponsorURL: {}'.format(sponsor_url))
                logging.info('\n' + str(pt))
            except KeyError as e:
                logging.error(e)
                exit(1)
        else:
            if filecmp.cmp(os.path.join(fork_repo, mirrors_dir, i), os.path.join(temp_dir, i), shallow=True):
                continue
            else:
                if os.system('mirrorbits edit {} -mirror-file {}'.format(i[:-5], os.path.abspath(i))) != 0:
                    logging.error('mirrorbits edit failed :(')
                    sys.exit(1)
                logging.info('update mirror: {}'.format(i[:-5]))
    for i in before_yaml_lst:
        if i not in yaml_lst:
            if os.system('mirrorbits remove {}'.format(i[:-5])) != 0:
                logging.error('mirrorbits remove failed :(')
                sys.exit(1)
            logging.info('remove mirror: {}'.format(i[:-5]))
    # clean temp files
    for i in before_yaml_lst:
        if os.system('rm {}'.format(os.path.join(temp_dir, i))) != 0:
            logging.error('remove temp file {} failed :('.format(os.path.join(temp_dir, i)))
            sys.exit(1)
        logging.info('remove temp file {}'.format(os.path.join(temp_dir, i)))
    time.sleep(n)


if __name__ == '__main__':
    with open('refresh_mirrors.yaml', 'r') as fp:
        repo_info = yaml.load(fp.read(), Loader=yaml.Loader)
    fork_url = repo_info['fork_url']
    fork_repo = fork_url.split('/')[-1].split('.')[0]
    mirrors_dir = repo_info['mirrors_dir']
    sleep_time = repo_info['sleep_time']
    # get remote repo code
    logging.info('get remote repo code')
    if os.system('git clone {}'.format(fork_url)) != 0:
        logging.error('git clone failed :(')
        sys.exit(1)
    if fork_repo not in os.listdir(os.getcwd()):
        logging.error('Error! directory {} does not exists. Check out whether failed in git clone.'.format(fork_repo))
    while True:
        sync_and_refresh(sleep_time)
